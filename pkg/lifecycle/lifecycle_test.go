package lifecycle_test

import (
	"errors"
	"time"

	"github.com/Nitroxaddict/vigil/pkg/container"
	"github.com/Nitroxaddict/vigil/pkg/lifecycle"
	t "github.com/Nitroxaddict/vigil/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// errClient is a minimal container.Client implementation whose GetContainer
// always fails. It mirrors the production behaviour where the Docker SDK
// returns &container.Container{} (with nil containerInfo) alongside an error.
type errClient struct {
	getErr error
}

func (c errClient) ListContainers(_ t.Filter) ([]t.Container, error) { return nil, nil }
func (c errClient) GetContainer(_ t.ContainerID) (t.Container, error) {
	return container.NewContainer(nil, nil), c.getErr
}
func (c errClient) StopContainer(_ t.Container, _ time.Duration) error      { return nil }
func (c errClient) StartContainer(_ t.Container) (t.ContainerID, error)     { return "", nil }
func (c errClient) StartContainerWithImage(_ t.Container, _ string) (t.ContainerID, error) {
	return "", nil
}
func (c errClient) RenameContainer(_ t.Container, _ string) error { return nil }
func (c errClient) IsContainerStale(_ t.Container, _ t.UpdateParams) (bool, t.ImageID, error) {
	return false, "", nil
}
func (c errClient) ExecuteCommand(_ t.ContainerID, _ string, _ int) (bool, error) {
	return false, nil
}
func (c errClient) RemoveImageByID(_ t.ImageID) error    { return nil }
func (c errClient) WarnOnHeadPullFailed(_ t.Container) bool { return false }

var _ = Describe("ExecutePostUpdateCommand", func() {
	When("GetContainer returns an error", func() {
		It("does not panic on the empty container returned alongside the error", func() {
			client := errClient{getErr: errors.New("docker inspect failed")}
			Expect(func() {
				lifecycle.ExecutePostUpdateCommand(client, t.ContainerID("abc123"))
			}).NotTo(Panic())
		})
	})
})
