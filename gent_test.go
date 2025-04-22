package gent

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	t.Run("teddy", func(t *testing.T) {
		req := require.New(t)

		set := NewSet[string]()
		req.NotNil(set.m)
		req.Empty(set.m)
		req.Equal(0, set.Len(), "empty == 0")
		req.Equal(set.Len(), set.Count(), "len == count")

		item := "teddy"
		req.False(set.Has(item), "no teddy yet")
		req.Equal(set.Has(item), set.Contains(item), "no teddy, Has == Contains")
		req.True(set.Add(item), "teddy added")
		req.Equal(1, set.Len(), "with teddy len is 1")
		req.Equal(set.Len(), set.Count(), "with teddy len+count are 1")
		req.False(set.Add(item), "teddy not added again")
		req.True(set.Has(item), "teddy definitely exists")
		req.True(set.Remove(item), "teddy was removed")
		req.False(set.Remove(item), "teddy not removed again")
	})

	t.Run("max", func(t *testing.T) {
		req := require.New(t)
		item := "max"

		set := NewSet(item)
		req.NotEmpty(set.m)
		req.True(set.Has(item), "max payne in the house")
		req.Equal(set.Has(item), set.Contains(item), "max, Has == Contains")

		req.Equal(1, set.Len(), "max == 1")
		set.Clear()
		req.Empty(set.m, "without max, world is empty")
		req.Equal(0, set.Len(), "only max isn't zero")
	})

	t.Run("Equal", func(t *testing.T) {
		req := require.New(t)

		items := []string{"a1", "z2", "b3", "y4"}
		set := NewSet(items...)
		req.True(set.Equal(NewSet(items...)), "same quad")
		req.False(set.Equal(NewSet(items[1:]...)), "so empty without the first")
		req.False(set.Equal(NewSet(append([]string{"1a"}, items[1:]...)...)), "swapped first item")
	})

	t.Run("ForEach stop", func(t *testing.T) {
		set := NewSet(3, 1, 3)
		counter := 0
		set.ForEach(func(_ int, stop func()) {
			counter++
			stop()
		})
		require.Equal(t, 1, counter, "immediately stopped")
	})

	reduce := func(items []string, removed string) []string {
		for i, each := range items {
			if each == removed {
				return append(items[0:i], items[i+1:]...)
			}
		}
		panic(fmt.Sprintf("Not found: %s", removed))
	}

	t.Run("ForEach", func(t *testing.T) {
		items := []string{"karl", "jackson", "johnny"}
		NewSet(items...).ForEach(func(s string, _ func()) {
			items = reduce(items, s)
		})
		require.Empty(t, items, "ForEach should've removed all items")
	})

	t.Run("ForEachAll", func(t *testing.T) {
		items := []string{"hans", "gunter", "gertrud"}
		NewSet(items...).ForEachAll(func(s string) {
			items = reduce(items, s)
		})
		require.Empty(t, items, "ForEachAll should've removed all items")
	})

	t.Run("ToSlice", func(t *testing.T) {
		set := NewSet("m1", "o2", "o2", "n3")
		sliced := set.ToSlice()
		sort.Strings(sliced)
		require.Equal(t, []string{"m1", "n3", "o2"}, sliced)
	})
}

func TestTri(t *testing.T) {
	req := require.New(t)
	req.Equal(13, Tri(13 < 14, 13, 14))
	req.Equal(14, Tri(14 < 13, 13, 14))
}

func TestMap(t *testing.T) {
	double := func(i int) int { return 2 * i }
	require.Equal(
		t,
		[]string{"2", "4", "8"},
		Map(
			Map([]int{1, 2, 4}, double),
			strconv.Itoa))
}

func ExampleMap() {
	fmt.Print(
		Map(
			[]int{1, 2, 4},
			func(i int) string {
				return fmt.Sprintf("item: %d", i)
			}))
	// Output: [item: 1 item: 2 item: 4]
}

func TestFilter(t *testing.T) {
	require.Equal(
		t,
		[]int{1, 2, 4},
		Filter(
			[]int{1, 2, 3, 4},
			func(i int) bool { return i%3 != 0 }))
	require.Nil(t, Filter([]int{1, 2, 3}, func(_ int) bool { return false }))
}

func ExampleFilter() {
	filterOdd := func(i int) bool { return i%2 != 0 }
	fmt.Print(
		Filter(
			[]int{1, 2, 3, 4, 5},
			filterOdd))
	// Output: [1 3 5]
}

func TestOrPanic2(t *testing.T) {
	req := require.New(t)
	req.Equal("wow", OrPanic2("wow", nil)(""))
	req.PanicsWithValue(
		"Message: killed. Error: turn.",
		func() { OrPanic2("", errors.New("turn"))("killed") })
}

func ExampleOrPanic2() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print(r)
		}
	}()

	divide := func(a, b int) (int, error) {
		if b == 0 {
			return 0, errors.New("can't divide with zero")
		}
		return a / b, nil
	}

	// The usual case where the call succeeds and you don't need to deal with the error.
	result := OrPanic2(divide(10, 2))("oops")
	fmt.Println(result)

	// Failing case where you still don't have to deal with the error.
	OrPanic2(divide(1, 0))("nope")

	// Output:
	// 5
	// Message: nope. Error: can't divide with zero.
}

func TestSnapshot(t *testing.T) {
	type tick struct {
		x int
		y int
	}
	type param struct {
		name string
		x    []tick
		o    []tick
		init bool
	}

	// Tested function.
	draw := func(x, o []tick) string {
		board := [][]rune{
			{' ', ' ', ' '},
			{' ', ' ', ' '},
			{' ', ' ', ' '},
		}
		mark := func(s []tick, r rune) {
			for _, each := range s {
				board[each.y][each.x] = r
			}
		}
		mark(x, 'x')
		mark(o, 'o')
		rendered := ""
		for y := 2; y >= 0; y-- {
			rendered += string(board[y]) + "\n"
		}
		return rendered
	}

	suite := NewSnapshotSuite("testdata/snapshots")
	run := func(p param) {
		t.Run(p.name, func(t *testing.T) {
			req := require.New(t)

			equal := func(expected, actual, message string) {
				req.Equal(expected, actual, message)
			}
			snapshot := suite.NewSnapshot(ToSafeFilename(p.name), !p.init, equal)
			req.Nil(snapshot.Run(draw(p.x, p.o)))
		})
	}

	run(
		param{
			name: "x wins",
			x:    []tick{{0, 0}, {1, 1}, {2, 2}},
			o:    []tick{{0, 1}, {1, 0}},
		},
	)

	run(
		param{
			name: "stalemate",
			x:    []tick{{0, 0}, {1, 1}, {2, 0}},
			o:    []tick{{2, 2}, {2, 1}, {0, 2}},
		},
	)
}

func TestNewOption(t *testing.T) {
	snapshot := Snapshot{
		Name:   "my-name-is-vibe",
		filep:  "/a/b/c.pdf",
		verify: true,
	}
	require.Equal(
		t,
		snapshot,
		NewOption(
			Snapshot{filep: snapshot.filep},
			func(s *Snapshot) { s.Name = snapshot.Name },
			func(s *Snapshot) { s.verify = snapshot.verify },
		))
}
