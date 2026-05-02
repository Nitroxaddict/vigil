package sorter_test

import (
	"sort"
	"testing"

	"github.com/Nitroxaddict/vigil/pkg/container"
	"github.com/Nitroxaddict/vigil/pkg/sorter"
	"github.com/Nitroxaddict/vigil/pkg/types"

	dt "github.com/docker/docker/api/types"
)

func makeContainer(id, created string) types.Container {
	return container.NewContainer(
		&dt.ContainerJSON{
			ContainerJSONBase: &dt.ContainerJSONBase{
				ID:      id,
				Created: created,
				Name:    id,
			},
		},
		nil,
	)
}

// On a malformed second timestamp, the pre-fix code re-checked t1's err
// (which was nil) and wrongly set t1 again instead of falling back for t2.
// With the fix, both fallbacks use time.Now(), so the malformed entry is
// the "later" one and sorts last — preserving cleanupExcessWatchtowers'
// keep-last semantics deterministically.
func TestByCreatedMalformedSecondTimestampSortsLast(t *testing.T) {
	good := makeContainer("good", "2020-01-01T00:00:00.000000000Z")
	bad := makeContainer("bad", "not-a-real-timestamp")

	cs := []types.Container{bad, good}
	sort.Sort(sorter.ByCreated(cs))

	if cs[len(cs)-1].ContainerInfo().ID != "bad" {
		t.Fatalf("malformed timestamp should sort last; got order: %s, %s",
			cs[0].ContainerInfo().ID, cs[1].ContainerInfo().ID)
	}
}

func TestByCreatedMalformedFirstTimestampSortsLast(t *testing.T) {
	good := makeContainer("good", "2020-01-01T00:00:00.000000000Z")
	bad := makeContainer("bad", "still-not-a-real-timestamp")

	cs := []types.Container{good, bad}
	sort.Sort(sorter.ByCreated(cs))

	if cs[len(cs)-1].ContainerInfo().ID != "bad" {
		t.Fatalf("malformed timestamp should sort last; got order: %s, %s",
			cs[0].ContainerInfo().ID, cs[1].ContainerInfo().ID)
	}
}

func TestByCreatedWellFormedTimestampsSortAscending(t *testing.T) {
	older := makeContainer("older", "2020-01-01T00:00:00.000000000Z")
	newer := makeContainer("newer", "2024-01-01T00:00:00.000000000Z")

	cs := []types.Container{newer, older}
	sort.Sort(sorter.ByCreated(cs))

	if cs[0].ContainerInfo().ID != "older" || cs[1].ContainerInfo().ID != "newer" {
		t.Fatalf("expected older,newer; got %s,%s",
			cs[0].ContainerInfo().ID, cs[1].ContainerInfo().ID)
	}
}
