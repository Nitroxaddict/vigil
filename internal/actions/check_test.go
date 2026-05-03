package actions_test

import (
	"time"

	"github.com/Nitroxaddict/vigil/internal/actions"
	"github.com/Nitroxaddict/vigil/pkg/filters"
	"github.com/Nitroxaddict/vigil/pkg/types"

	. "github.com/Nitroxaddict/vigil/internal/actions/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckForSanity", func() {
	When("rolling restarts are disabled", func() {
		// Documents the rollingRestarts=false no-op contract that the
		// monitor-only call-site in cmd/root.go relies on (ATL-36):
		// CheckForSanity must return nil without inspecting containers,
		// so a linked container is irrelevant.
		It("should be a no-op even with linked containers", func() {
			linked := CreateMockContainerWithLinks(
				"linked-container",
				"/linked-container",
				"fake-image:latest",
				time.Now(),
				[]string{"other-container"},
				CreateMockImageInfo("fake-image:latest"),
			)
			client := CreateMockClient(
				&TestData{
					Containers: []types.Container{linked},
				},
				false,
				false,
			)

			Expect(actions.CheckForSanity(client, filters.NoFilter, false)).To(Succeed())
		})
	})
})
