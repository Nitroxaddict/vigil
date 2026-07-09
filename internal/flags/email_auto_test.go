package flags

// Tests for the email notifier auto-default logic in ProcessFlagAliases:
// when email is the sole legacy notifier, notification-template defaults to
// email-html (or email-text when --notification-email-plaintext is set) and
// notification-report is forced to true.

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEmailFlagsCmd(t *testing.T) *cobra.Command {
	t.Helper()
	logrus.StandardLogger().ExitFunc = func(_ int) { t.FailNow() }
	cmd := new(cobra.Command)
	SetDefaults()
	RegisterDockerFlags(cmd)
	RegisterSystemFlags(cmd)
	RegisterNotificationFlags(cmd)
	return cmd
}

func TestEmailAutoDefault_HTMLWhenEmailSoleNotifier(t *testing.T) {
	cmd := newEmailFlagsCmd(t)
	require.NoError(t, cmd.ParseFlags([]string{"--notifications", "email"}))
	ProcessFlagAliases(cmd.Flags())

	tpl, _ := cmd.Flags().GetString("notification-template")
	assert.Equal(t, "email-html", tpl, "notification-template must default to email-html when email is sole notifier")

	report, _ := cmd.Flags().GetBool("notification-report")
	assert.True(t, report, "notification-report must be true when email-html is auto-defaulted")
}

func TestEmailAutoDefault_TextWhenPlaintextFlagSet(t *testing.T) {
	cmd := newEmailFlagsCmd(t)
	require.NoError(t, cmd.ParseFlags([]string{
		"--notifications", "email",
		"--notification-email-plaintext",
	}))
	ProcessFlagAliases(cmd.Flags())

	tpl, _ := cmd.Flags().GetString("notification-template")
	assert.Equal(t, "email-text", tpl, "notification-template must be email-text when --notification-email-plaintext is set")
}

func TestEmailAutoDefault_ExplicitTemplateNotOverridden(t *testing.T) {
	cmd := newEmailFlagsCmd(t)
	require.NoError(t, cmd.ParseFlags([]string{
		"--notifications", "email",
		"--notification-template", "default",
	}))
	ProcessFlagAliases(cmd.Flags())

	tpl, _ := cmd.Flags().GetString("notification-template")
	assert.Equal(t, "default", tpl, "explicit --notification-template must not be overridden")
}

func TestEmailAutoDefault_NotFiredWhenMultipleNotifiers(t *testing.T) {
	cmd := newEmailFlagsCmd(t)
	require.NoError(t, cmd.ParseFlags([]string{
		"--notifications", "email,slack",
	}))
	ProcessFlagAliases(cmd.Flags())

	tpl, _ := cmd.Flags().GetString("notification-template")
	assert.NotEqual(t, "email-html", tpl, "auto-default must not fire when multiple notifiers are configured")
}

func TestEmailAutoDefault_NotFiredWhenNotificationURLPresent(t *testing.T) {
	cmd := newEmailFlagsCmd(t)
	require.NoError(t, cmd.ParseFlags([]string{
		"--notifications", "email",
		"--notification-url", "slack://hook:tokenA-tokenB-tokenC@webhook",
	}))
	ProcessFlagAliases(cmd.Flags())

	tpl, _ := cmd.Flags().GetString("notification-template")
	assert.NotEqual(t, "email-html", tpl,
		"auto-default must not fire when --notification-url entries are present (HTML would leak into other notifiers)")
}
