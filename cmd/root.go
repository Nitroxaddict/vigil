package cmd

import (
	"errors"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Nitroxaddict/vigil/internal/actions"
	"github.com/Nitroxaddict/vigil/internal/flags"
	"github.com/Nitroxaddict/vigil/internal/meta"
	"github.com/Nitroxaddict/vigil/pkg/api"
	apiMetrics "github.com/Nitroxaddict/vigil/pkg/api/metrics"
	"github.com/Nitroxaddict/vigil/pkg/api/update"
	"github.com/Nitroxaddict/vigil/pkg/container"
	"github.com/Nitroxaddict/vigil/pkg/filters"
	"github.com/Nitroxaddict/vigil/pkg/metrics"
	"github.com/Nitroxaddict/vigil/pkg/notifications"
	t "github.com/Nitroxaddict/vigil/pkg/types"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	client            container.Client
	scheduleSpec      string
	cleanup           bool
	noRestart         bool
	noPull            bool
	monitorOnly       bool
	enableLabel       bool
	disableContainers []string
	notifier          t.Notifier
	timeout           time.Duration
	lifecycleHooks    bool
	batchRestart      bool
	scope             string
	labelPrecedence   bool
)

var rootCmd = NewRootCommand()

// NewRootCommand creates the root command for vigil
func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "vigil",
		Short: "Automatically updates running Docker containers",
		Long: `
	Vigil automatically updates running Docker containers whenever a new image is released.
	A community-maintained fork of Watchtower. https://github.com/Nitroxaddict/vigil
	`,
		Run:    Run,
		PreRun: PreRun,
		Args:   cobra.ArbitraryArgs,
	}
}

func init() {
	flags.SetDefaults()
	flags.RegisterDockerFlags(rootCmd)
	flags.RegisterSystemFlags(rootCmd)
	flags.RegisterNotificationFlags(rootCmd)
}

// Execute the root func and exit in case of errors
func Execute() {
	rootCmd.AddCommand(notifyUpgradeCommand)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// PreRun is a lifecycle hook that runs before the command is executed.
func PreRun(cmd *cobra.Command, _ []string) {
	f := cmd.PersistentFlags()
	flags.ProcessFlagAliases(f)
	if err := flags.SetupLogging(f); err != nil {
		log.Fatalf("Failed to initialize logging: %s", err.Error())
	}

	scheduleSpec, _ = f.GetString("schedule")

	flags.GetSecretsFromFiles(cmd)
	cleanup, noRestart, monitorOnly, timeout = flags.ReadFlags(cmd)

	if timeout < 0 {
		log.Fatal("Please specify a positive value for timeout value.")
	}

	enableLabel, _ = f.GetBool("label-enable")
	disableContainers, _ = f.GetStringSlice("disable-containers")
	lifecycleHooks, _ = f.GetBool("enable-lifecycle-hooks")
	batchRestart, _ = f.GetBool("batch-restart")
	if f.Changed("rolling-restart") {
		log.Warn("--rolling-restart is deprecated and will be removed in v2. Rolling restart is now the default. Remove this flag from your configuration.")
	}
	if os.Getenv("WATCHTOWER_ROLLING_RESTART") != "" || os.Getenv("VIGIL_ROLLING_RESTART") != "" {
		log.Warn("WATCHTOWER_ROLLING_RESTART / VIGIL_ROLLING_RESTART environment variables are deprecated and ignored. Rolling restart is now the default.")
	}
	scope, _ = f.GetString("scope")
	labelPrecedence, _ = f.GetBool("label-take-precedence")

	if scope != "" {
		log.Debugf(`Using scope %q`, scope)
	}

	// configure environment vars for client
	err := flags.EnvConfig(cmd)
	if err != nil {
		log.Fatal(err)
	}

	noPull, _ = f.GetBool("no-pull")
	includeStopped, _ := f.GetBool("include-stopped")
	includeRestarting, _ := f.GetBool("include-restarting")
	reviveStopped, _ := f.GetBool("revive-stopped")
	removeVolumes, _ := f.GetBool("remove-volumes")
	warnOnHeadPullFailed, _ := f.GetString("warn-on-head-failure")

	if monitorOnly && noPull {
		log.Warn("Using NO_PULL and MONITOR_ONLY simultaneously might lead to no action being taken at all. If this is intentional, you may safely ignore this message.")
	}

	client = container.NewClient(container.ClientOptions{
		IncludeStopped:    includeStopped,
		ReviveStopped:     reviveStopped,
		RemoveVolumes:     removeVolumes,
		IncludeRestarting: includeRestarting,
		WarnOnHeadFailed:  container.WarningStrategy(warnOnHeadPullFailed),
	})

	notifier = notifications.NewNotifier(cmd)
	notifier.AddLogHook()
}

// Run is the main execution flow of the command
func Run(c *cobra.Command, names []string) {
	filter, filterDesc := filters.BuildFilter(names, disableContainers, enableLabel, scope)
	runOnce, _ := c.PersistentFlags().GetBool("run-once")
	enableUpdateAPI, _ := c.PersistentFlags().GetBool("http-api-update")
	enableMetricsAPI, _ := c.PersistentFlags().GetBool("http-api-metrics")
	unblockHTTPAPI, _ := c.PersistentFlags().GetBool("http-api-periodic-polls")
	apiToken, _ := c.PersistentFlags().GetString("http-api-token")
	healthCheck, _ := c.PersistentFlags().GetBool("health-check")

	if healthCheck {
		// health check should not have pid 1
		if os.Getpid() == 1 {
			time.Sleep(1 * time.Second)
			log.Fatal("The health check flag should never be passed to the main vigil container process")
		}
		os.Exit(0)
	}

	awaitDockerClient()

	if err := actions.CheckForSanity(client, filter, !batchRestart && !monitorOnly); err != nil {
		logNotifyExit(err)
	}

	if runOnce {
		writeStartupMessage(c, time.Time{}, filterDesc)
		runUpdatesWithNotifications(filter)
		notifier.Close()
		os.Exit(0)
		return
	}

	if err := actions.CheckForMultipleWatchtowerInstances(client, cleanup, scope); err != nil {
		logNotifyExit(err)
	}

	// The lock is shared between the scheduler and the HTTP API. It only allows one update to run at a time.
	updateLock := make(chan bool, 1)
	updateLock <- true

	httpAPI := api.New(apiToken)

	if enableUpdateAPI {
		updateHandler := update.New(func(images []string) {
			metric := runUpdatesWithNotifications(filters.FilterByImage(images, filter))
			metrics.RegisterScan(metric)
		}, updateLock)
		httpAPI.RegisterFunc(updateHandler.Path, updateHandler.Handle)
		// If polling isn't enabled the scheduler is never started, and
		// we need to trigger the startup messages manually.
		if !unblockHTTPAPI {
			writeStartupMessage(c, time.Time{}, filterDesc)
		}
	}

	if enableMetricsAPI {
		metricsHandler := apiMetrics.New()
		httpAPI.RegisterHandler(metricsHandler.Path, metricsHandler.Handle)
	}

	if err := httpAPI.Start(enableUpdateAPI && !unblockHTTPAPI); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start API", err)
	}

	if err := runUpgradesOnSchedule(c, filter, filterDesc, updateLock); err != nil {
		log.Error(err)
	}

	os.Exit(1)
}

func logNotifyExit(err error) {
	log.Error(err)
	notifier.Close()
	os.Exit(1)
}

func awaitDockerClient() {
	log.Debug("Sleeping for a second to ensure the docker api client has been properly initialized.")
	time.Sleep(1 * time.Second)
}

func formatDuration(d time.Duration) string {
	totalHours := int64(d.Hours())
	days := totalHours / 24
	hours := totalHours % 24
	minutes := int64(math.Mod(d.Minutes(), 60))
	seconds := int64(math.Mod(d.Seconds(), 60))

	var parts []string

	if days == 1 {
		parts = append(parts, "1 day")
	} else if days != 0 {
		parts = append(parts, strconv.FormatInt(days, 10)+" days")
	}

	if hours == 1 {
		parts = append(parts, "1 hour")
	} else if hours != 0 {
		parts = append(parts, strconv.FormatInt(hours, 10)+" hours")
	}

	if minutes == 1 {
		parts = append(parts, "1 minute")
	} else if minutes != 0 {
		parts = append(parts, strconv.FormatInt(minutes, 10)+" minutes")
	}

	if seconds == 1 {
		parts = append(parts, "1 second")
	} else if seconds != 0 {
		parts = append(parts, strconv.FormatInt(seconds, 10)+" seconds")
	}

	if len(parts) == 0 {
		return "0 seconds"
	}

	return strings.Join(parts, ", ")
}

func writeStartupMessage(c *cobra.Command, sched time.Time, filtering string) {
	noStartupMessage, _ := c.PersistentFlags().GetBool("no-startup-message")
	enableUpdateAPI, _ := c.PersistentFlags().GetBool("http-api-update")

	var startupLog *log.Entry
	if noStartupMessage {
		startupLog = notifications.LocalLog
	} else {
		startupLog = log.NewEntry(log.StandardLogger())
		// Batch up startup messages to send them as a single notification
		notifier.StartNotification()
	}

	startupLog.Info("Vigil ", meta.Version)

	notifierNames := notifier.GetNames()
	if len(notifierNames) > 0 {
		startupLog.Info("Using notifications: " + strings.Join(notifierNames, ", "))
	} else {
		startupLog.Info("Using no notifications")
	}

	startupLog.Info(filtering)

	if !sched.IsZero() {
		until := formatDuration(time.Until(sched))
		startupLog.Info("Scheduling first run: " + sched.Format("2006-01-02 15:04:05 -0700 MST"))
		startupLog.Info("Note that the first check will be performed in " + until)
	} else if runOnce, _ := c.PersistentFlags().GetBool("run-once"); runOnce {
		startupLog.Info("Running a one time update.")
	} else {
		startupLog.Info("Periodic runs are not enabled.")
	}

	if enableUpdateAPI {
		// TODO: make listen port configurable
		startupLog.Info("The HTTP API is enabled at :8080.")
	}

	if !noStartupMessage {
		// Send the queued up startup messages, not including the trace warning below (to make sure it's noticed)
		notifier.SendNotification(nil)
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		startupLog.Warn("Trace level enabled: log will include sensitive information as credentials and tokens")
	}
}

// scheduleTickInterval is how often the schedule loop wakes up to compare the
// wall clock against the next scheduled run. It bounds how late a fire can be
// after wake from host sleep / Docker VM pause; daily-cadence schedules don't
// need finer resolution.
const scheduleTickInterval = 30 * time.Second

func runUpgradesOnSchedule(c *cobra.Command, filter t.Filter, filtering string, lock chan bool) error {
	if lock == nil {
		lock = make(chan bool, 1)
		lock <- true
	}

	schedule, err := cron.Parse(scheduleSpec)
	if err != nil {
		return err
	}

	firstRun := schedule.Next(time.Now())
	writeStartupMessage(c, firstRun, filtering)

	fire := func(now time.Time) {
		select {
		case v := <-lock:
			defer func() { lock <- v }()
			metric := runUpdatesWithNotifications(filter)
			metrics.RegisterScan(metric)
		default:
			// Update was skipped
			metrics.RegisterScan(nil)
			log.Debug("Skipped another update already running.")
		}
		log.Debug("Scheduled next run: " + schedule.Next(now).String())
	}

	// Graceful shut-down on SIGINT/SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, syscall.SIGTERM)
	stop := make(chan struct{})
	go func() {
		<-interrupt
		close(stop)
	}()

	ticker := time.NewTicker(scheduleTickInterval)
	defer ticker.Stop()

	runScheduleLoop(schedule, firstRun, ticker.C, stop, fire)

	log.Info("Waiting for running update to be finished...")
	<-lock
	return nil
}

// runScheduleLoop fires `fire` whenever a tick arrives at or after `next`.
// Decisions are made on each tick's wall-clock value, never on monotonic time,
// so the loop survives host sleep / Docker Desktop VM pause: when the runtime
// resumes after missing one or more windows, the next tick sees now >= next,
// fires once for the most recent missed window, then advances `next` to the
// first future window per `schedule`. Multiple windows missed during a single
// sleep collapse into a single fire — appropriate for a container updater
// (pull latest now, do not replay history).
func runScheduleLoop(
	schedule cron.Schedule,
	next time.Time,
	tick <-chan time.Time,
	stop <-chan struct{},
	fire func(time.Time),
) {
	for {
		select {
		case now, ok := <-tick:
			if !ok {
				return
			}
			if !now.Before(next) {
				fire(now)
				next = schedule.Next(now)
			}
		case <-stop:
			return
		}
	}
}

func runUpdatesWithNotifications(filter t.Filter) *metrics.Metric {
	notifier.StartNotification()
	updateParams := t.UpdateParams{
		Filter:          filter,
		Cleanup:         cleanup,
		NoRestart:       noRestart,
		Timeout:         timeout,
		MonitorOnly:     monitorOnly,
		LifecycleHooks:  lifecycleHooks,
		RollingRestart:  !batchRestart,
		LabelPrecedence: labelPrecedence,
		NoPull:          noPull,
	}
	result, err := actions.Update(client, updateParams)
	if err != nil {
		log.Error(err)
	}
	notifier.SendNotification(result)
	metricResults := metrics.NewMetric(result)
	notifications.LocalLog.WithFields(log.Fields{
		"Scanned": metricResults.Scanned,
		"Updated": metricResults.Updated,
		"Failed":  metricResults.Failed,
	}).Info("Session done")
	return metricResults
}
