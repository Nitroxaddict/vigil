package container

import (
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	log "github.com/sirupsen/logrus"
)

// apiVersionAtLeast returns true if the given API version string (e.g. "1.43")
// is greater than or equal to the specified major.minor version.
func apiVersionAtLeast(apiVersion string, major, minor int) bool {
	parts := strings.SplitN(apiVersion, ".", 2)
	if len(parts) != 2 {
		return false
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	if maj != major {
		return maj > major
	}
	return min >= minor
}

// sanitizeNetworkConfig strips fields from EndpointSettings that require a
// higher Docker API version than what the daemon supports.
func sanitizeNetworkConfig(config *network.NetworkingConfig, apiVersion string) {
	if config == nil {
		return
	}

	if !apiVersionAtLeast(apiVersion, 1, 44) {
		for name, ep := range config.EndpointsConfig {
			if ep.MacAddress != "" {
				log.Debugf("Stripping per-network MacAddress from %s (requires API 1.44, have %s)", name, apiVersion)
				ep.MacAddress = ""
			}
		}
	}
}

// sanitizeContainerConfig strips fields from container.Config that require a
// higher Docker API version than what the daemon supports.
func sanitizeContainerConfig(config *container.Config, apiVersion string) {
	if config == nil {
		return
	}

	if !apiVersionAtLeast(apiVersion, 1, 44) {
		if config.Healthcheck != nil && config.Healthcheck.StartInterval != 0 {
			log.Debugf("Stripping HealthConfig.StartInterval (requires API 1.44, have %s)", apiVersion)
			config.Healthcheck.StartInterval = 0
		}
	}
}

// sanitizeConfigs is a convenience function that sanitizes all config types
// based on the daemon's API version. It logs once at info level when any
// sanitization is performed.
func sanitizeConfigs(containerConfig *container.Config, networkConfig *network.NetworkingConfig, apiVersion string) {
	if apiVersion == "" {
		log.Warn("Unknown Docker API version, skipping config sanitization")
		return
	}

	if !apiVersionAtLeast(apiVersion, 1, 44) {
		log.Debugf("Docker API version %s < 1.44, sanitizing config fields", apiVersion)
	}

	sanitizeContainerConfig(containerConfig, apiVersion)
	sanitizeNetworkConfig(networkConfig, apiVersion)
}
