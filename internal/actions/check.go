package actions

import (
	"fmt"
	"sort"
	"time"

	"github.com/Nitroxaddict/vigil/pkg/container"
	"github.com/Nitroxaddict/vigil/pkg/filters"
	"github.com/Nitroxaddict/vigil/pkg/sorter"
	"github.com/Nitroxaddict/vigil/pkg/types"

	log "github.com/sirupsen/logrus"
)

// CheckForSanity makes sure everything is sane before starting
func CheckForSanity(client container.Client, filter types.Filter, rollingRestarts bool) error {
	log.Debug("Making sure everything is sane before starting")

	if rollingRestarts {
		containers, err := client.ListContainers(filter)
		if err != nil {
			return err
		}
		for _, c := range containers {
			if len(c.Links()) > 0 {
				return fmt.Errorf(
					"%q is depending on at least one other container. This is not compatible with rolling restarts",
					c.Name(),
				)
			}
		}
	}
	return nil
}

// CheckForMultipleWatchtowerInstances will ensure that there are not multiple instances of the
// updater running simultaneously. If multiple vigil containers are detected, this function
// will stop and remove all but the most recently started container. This behaviour can be bypassed
// if a scope UID is defined.
func CheckForMultipleWatchtowerInstances(client container.Client, cleanup bool, scope string) error {
	filter := filters.WatchtowerContainersFilter
	if scope != "" {
		filter = filters.FilterByScope(scope, filter)
	}
	containers, err := client.ListContainers(filter)

	if err != nil {
		return err
	}

	if len(containers) <= 1 {
		log.Debug("There are no additional vigil containers")
		return nil
	}

	log.Info("Found multiple running vigil instances. Cleaning up.")
	return cleanupExcessWatchtowers(containers, client, cleanup)
}

func cleanupExcessWatchtowers(containers []types.Container, client container.Client, cleanup bool) error {
	var stopErrors int

	sort.Sort(sorter.ByCreated(containers))
	allContainersExceptLast := containers[0 : len(containers)-1]

	for _, c := range allContainersExceptLast {
		if err := client.StopContainer(c, 10*time.Minute); err != nil {
			// logging the original here as we're just returning a count
			log.WithError(err).Error("Could not stop a previous vigil instance.")
			stopErrors++
			continue
		}

		if cleanup {
			if err := client.RemoveImageByID(c.ImageID()); err != nil {
				log.WithError(err).Warning("Could not cleanup vigil images, possibly because of other vigil instances in other scopes.")
			}
		}
	}

	if stopErrors > 0 {
		return fmt.Errorf("%d errors while stopping vigil containers", stopErrors)
	}

	return nil
}
