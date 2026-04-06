package container

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
)

func TestApiVersionAtLeast(t *testing.T) {
	tests := []struct {
		version  string
		major    int
		minor    int
		expected bool
	}{
		{"1.43", 1, 44, false},
		{"1.44", 1, 44, true},
		{"1.45", 1, 44, true},
		{"1.43", 1, 43, true},
		{"2.0", 1, 44, true},
		{"0.99", 1, 44, false},
		{"", 1, 44, false},
		{"invalid", 1, 44, false},
		{"1.x", 1, 44, false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := apiVersionAtLeast(tt.version, tt.major, tt.minor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeNetworkConfig(t *testing.T) {
	t.Run("strips MacAddress when API < 1.44", func(t *testing.T) {
		config := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"bridge": {MacAddress: "02:42:ac:11:00:02"},
				"custom": {MacAddress: "02:42:ac:11:00:03", Aliases: []string{"app"}},
			},
		}
		sanitizeNetworkConfig(config, "1.43")
		assert.Equal(t, "", config.EndpointsConfig["bridge"].MacAddress)
		assert.Equal(t, "", config.EndpointsConfig["custom"].MacAddress)
		assert.Equal(t, []string{"app"}, config.EndpointsConfig["custom"].Aliases)
	})

	t.Run("preserves MacAddress when API >= 1.44", func(t *testing.T) {
		config := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"bridge": {MacAddress: "02:42:ac:11:00:02"},
			},
		}
		sanitizeNetworkConfig(config, "1.44")
		assert.Equal(t, "02:42:ac:11:00:02", config.EndpointsConfig["bridge"].MacAddress)
	})

	t.Run("handles nil config", func(t *testing.T) {
		sanitizeNetworkConfig(nil, "1.43") // should not panic
	})

	t.Run("handles empty endpoints", func(t *testing.T) {
		config := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{},
		}
		sanitizeNetworkConfig(config, "1.43") // should not panic
	})
}

func TestSanitizeContainerConfig(t *testing.T) {
	t.Run("strips StartInterval when API < 1.44", func(t *testing.T) {
		config := &container.Config{
			Healthcheck: &container.HealthConfig{
				Test:          []string{"CMD", "curl", "-f", "http://localhost/"},
				Interval:      30 * time.Second,
				StartInterval: 5 * time.Second,
			},
		}
		sanitizeContainerConfig(config, "1.43")
		assert.Equal(t, time.Duration(0), config.Healthcheck.StartInterval)
		// Other healthcheck fields should be preserved
		assert.Equal(t, 30*time.Second, config.Healthcheck.Interval)
		assert.Equal(t, []string{"CMD", "curl", "-f", "http://localhost/"}, config.Healthcheck.Test)
	})

	t.Run("preserves StartInterval when API >= 1.44", func(t *testing.T) {
		config := &container.Config{
			Healthcheck: &container.HealthConfig{
				StartInterval: 5 * time.Second,
			},
		}
		sanitizeContainerConfig(config, "1.44")
		assert.Equal(t, 5*time.Second, config.Healthcheck.StartInterval)
	})

	t.Run("handles nil config", func(t *testing.T) {
		sanitizeContainerConfig(nil, "1.43") // should not panic
	})

	t.Run("handles nil healthcheck", func(t *testing.T) {
		config := &container.Config{}
		sanitizeContainerConfig(config, "1.43") // should not panic
	})

	t.Run("handles zero StartInterval gracefully", func(t *testing.T) {
		config := &container.Config{
			Healthcheck: &container.HealthConfig{
				StartInterval: 0,
			},
		}
		sanitizeContainerConfig(config, "1.43") // should not panic or log
		assert.Equal(t, time.Duration(0), config.Healthcheck.StartInterval)
	})
}
