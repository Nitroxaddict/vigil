package container

import "strconv"

const (
	// Vigil labels (preferred)
	vigilLabel                    = "dev.vigil"
	vigilSignalLabel              = "dev.vigil.stop-signal"
	vigilEnableLabel              = "dev.vigil.enable"
	vigilMonitorOnlyLabel         = "dev.vigil.monitor-only"
	vigilNoPullLabel              = "dev.vigil.no-pull"
	vigilDependsOnLabel           = "dev.vigil.depends-on"
	vigilScopeLabel               = "dev.vigil.scope"
	vigilPreCheckLabel            = "dev.vigil.lifecycle.pre-check"
	vigilPostCheckLabel           = "dev.vigil.lifecycle.post-check"
	vigilPreUpdateLabel           = "dev.vigil.lifecycle.pre-update"
	vigilPostUpdateLabel          = "dev.vigil.lifecycle.post-update"
	vigilPreUpdateTimeoutLabel    = "dev.vigil.lifecycle.pre-update-timeout"
	vigilPostUpdateTimeoutLabel   = "dev.vigil.lifecycle.post-update-timeout"

	// Legacy Watchtower labels (still supported for backward compatibility)
	watchtowerLabel        = "com.centurylinklabs.watchtower"
	signalLabel            = "com.centurylinklabs.watchtower.stop-signal"
	enableLabel            = "com.centurylinklabs.watchtower.enable"
	monitorOnlyLabel       = "com.centurylinklabs.watchtower.monitor-only"
	noPullLabel            = "com.centurylinklabs.watchtower.no-pull"
	dependsOnLabel         = "com.centurylinklabs.watchtower.depends-on"
	scope                  = "com.centurylinklabs.watchtower.scope"
	preCheckLabel          = "com.centurylinklabs.watchtower.lifecycle.pre-check"
	postCheckLabel         = "com.centurylinklabs.watchtower.lifecycle.post-check"
	preUpdateLabel         = "com.centurylinklabs.watchtower.lifecycle.pre-update"
	postUpdateLabel        = "com.centurylinklabs.watchtower.lifecycle.post-update"
	preUpdateTimeoutLabel  = "com.centurylinklabs.watchtower.lifecycle.pre-update-timeout"
	postUpdateTimeoutLabel = "com.centurylinklabs.watchtower.lifecycle.post-update-timeout"
)

// vigilLabelMap maps each legacy Watchtower label to its Vigil equivalent.
// Vigil labels take precedence when both are present.
var vigilLabelMap = map[string]string{
	watchtowerLabel:        vigilLabel,
	signalLabel:            vigilSignalLabel,
	enableLabel:            vigilEnableLabel,
	monitorOnlyLabel:       vigilMonitorOnlyLabel,
	noPullLabel:            vigilNoPullLabel,
	dependsOnLabel:         vigilDependsOnLabel,
	scope:                  vigilScopeLabel,
	preCheckLabel:          vigilPreCheckLabel,
	postCheckLabel:         vigilPostCheckLabel,
	preUpdateLabel:         vigilPreUpdateLabel,
	postUpdateLabel:        vigilPostUpdateLabel,
	preUpdateTimeoutLabel:  vigilPreUpdateTimeoutLabel,
	postUpdateTimeoutLabel: vigilPostUpdateTimeoutLabel,
}

// GetLifecyclePreCheckCommand returns the pre-check command set in the container metadata or an empty string
func (c Container) GetLifecyclePreCheckCommand() string {
	return c.getLabelValueOrEmpty(preCheckLabel)
}

// GetLifecyclePostCheckCommand returns the post-check command set in the container metadata or an empty string
func (c Container) GetLifecyclePostCheckCommand() string {
	return c.getLabelValueOrEmpty(postCheckLabel)
}

// GetLifecyclePreUpdateCommand returns the pre-update command set in the container metadata or an empty string
func (c Container) GetLifecyclePreUpdateCommand() string {
	return c.getLabelValueOrEmpty(preUpdateLabel)
}

// GetLifecyclePostUpdateCommand returns the post-update command set in the container metadata or an empty string
func (c Container) GetLifecyclePostUpdateCommand() string {
	return c.getLabelValueOrEmpty(postUpdateLabel)
}

// ContainsWatchtowerLabel takes a map of labels and values and tells
// the consumer whether it contains a valid watchtower/vigil instance label
func ContainsWatchtowerLabel(labels map[string]string) bool {
	if val, ok := labels[vigilLabel]; ok && val == "true" {
		return true
	}
	val, ok := labels[watchtowerLabel]
	return ok && val == "true"
}

// getLabelValueOrEmpty checks for the Vigil label first, then falls back to the legacy label
func (c Container) getLabelValueOrEmpty(label string) string {
	if vigilEquiv, ok := vigilLabelMap[label]; ok {
		if val, ok := c.containerInfo.Config.Labels[vigilEquiv]; ok {
			return val
		}
	}
	if val, ok := c.containerInfo.Config.Labels[label]; ok {
		return val
	}
	return ""
}

// getLabelValue checks for the Vigil label first, then falls back to the legacy label
func (c Container) getLabelValue(label string) (string, bool) {
	if vigilEquiv, ok := vigilLabelMap[label]; ok {
		if val, ok := c.containerInfo.Config.Labels[vigilEquiv]; ok {
			return val, true
		}
	}
	val, ok := c.containerInfo.Config.Labels[label]
	return val, ok
}

func (c Container) getBoolLabelValue(label string) (bool, error) {
	if vigilEquiv, ok := vigilLabelMap[label]; ok {
		if strVal, ok := c.containerInfo.Config.Labels[vigilEquiv]; ok {
			value, err := strconv.ParseBool(strVal)
			return value, err
		}
	}
	if strVal, ok := c.containerInfo.Config.Labels[label]; ok {
		value, err := strconv.ParseBool(strVal)
		return value, err
	}
	return false, errorLabelNotFound
}
