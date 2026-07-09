package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wt "github.com/Nitroxaddict/vigil/pkg/types"
)

func makeStatusWithLabels(labels map[string]string) *ContainerStatus {
	return &ContainerStatus{
		containerID:    wt.ContainerID("abc123"),
		newImage:       wt.ImageID("sha256:deadbeef1234"),
		newImageLabels: labels,
	}
}

func TestLatestImageVersion_absent(t *testing.T) {
	s := makeStatusWithLabels(map[string]string{
		"other.label": "v1.0",
	})
	assert.Equal(t, "", s.LatestImageVersion(), "absent OCI version label must return empty string")
}

func TestLatestImageVersion_nilLabels(t *testing.T) {
	s := makeStatusWithLabels(nil)
	require.NotPanics(t, func() {
		v := s.LatestImageVersion()
		assert.Equal(t, "", v, "nil label map must return empty string without panic")
	})
}

func TestLatestImageVersion_present(t *testing.T) {
	s := makeStatusWithLabels(map[string]string{
		"org.opencontainers.image.version": "1.45.3",
	})
	assert.Equal(t, "1.45.3", s.LatestImageVersion())
}

func TestLatestImageVersion_truncatesLongLabel(t *testing.T) {
	// 1 MB label value must be truncated to maxLabelLen bytes
	big := strings.Repeat("a", 1024*1024)
	s := makeStatusWithLabels(map[string]string{
		"org.opencontainers.image.version": big,
	})
	v := s.LatestImageVersion()
	assert.LessOrEqual(t, len(v), maxLabelLen, "label must be truncated to maxLabelLen bytes")
}

func TestLatestImageVersion_stripsControlChars(t *testing.T) {
	// CR, LF, NUL are control characters; they must be stripped.
	raw := "1.0\r\n\x00injection"
	s := makeStatusWithLabels(map[string]string{
		"org.opencontainers.image.version": raw,
	})
	v := s.LatestImageVersion()
	assert.NotContains(t, v, "\r", "CR must be stripped from version label")
	assert.NotContains(t, v, "\n", "LF must be stripped from version label")
	assert.NotContains(t, v, "\x00", "NUL must be stripped from version label")
	assert.Contains(t, v, "1.0", "printable content must be preserved")
}

func TestSanitizeLabel_truncatesExact(t *testing.T) {
	// exactly maxLabelLen chars — no truncation
	s := strings.Repeat("x", maxLabelLen)
	assert.Equal(t, s, sanitizeLabel(s))

	// maxLabelLen+1 — truncated
	long := strings.Repeat("x", maxLabelLen+1)
	result := sanitizeLabel(long)
	assert.Equal(t, maxLabelLen, len(result))
}
