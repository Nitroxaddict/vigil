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

// Both-malformed is the case the single `now := time.Now()` capture in
// Less exists for. Pre-capture, t1 and t2 each got their own time.Now()
// (with t1's call microseconds before t2's), so t1.Before(t2) was always
// true — Less became non-deterministic across pair comparisons during
// sort.Sort, violating the strict-weak-ordering contract. Existing
// internal/actions tests use time.Time.String() (not RFC3339Nano) for
// mock containers, so this is the realistic hot path, not an edge case.
func TestByCreatedBothMalformedTimestampsAreNotStrictlyOrdered(t *testing.T) {
	a := makeContainer("a", "not-a-timestamp")
	b := makeContainer("b", "also-not-a-timestamp")
	cs := sorter.ByCreated{a, b}

	if cs.Less(0, 1) || cs.Less(1, 0) {
		t.Fatalf("Less must return false both ways when both timestamps are malformed; got Less(0,1)=%v Less(1,0)=%v",
			cs.Less(0, 1), cs.Less(1, 0))
	}
}
