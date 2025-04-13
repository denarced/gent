package gent

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	req := require.New(t)

	set := NewSet[string]()
	req.NotNil(set.m)
	req.Empty(set.m)
	req.Equal(0, set.Len(), "empty == 0")

	item := "teddy"
	req.False(set.Has(item), "no teddy yet")
	req.True(set.Add(item), "teddy added")
	req.Equal(1, set.Len(), "with teddy len is 1")
	req.False(set.Add(item), "teddy not added again")
	req.True(set.Has(item), "teddy definitely exists")
	req.True(set.Remove(item), "teddy was removed")
	req.False(set.Remove(item), "teddy not removed again")

	item = "max"
	set = NewSet(item)
	req.NotEmpty(set.m)
	req.True(set.Has(item), "max payne in the house")

	req.Equal(1, set.Len(), "max == 1")
	set.Clear()
	req.Empty(set.m, "without max, world is empty")
	req.Equal(0, set.Len(), "only max isn't zero")

	items := []string{"a1", "z2", "b3", "y4"}
	set = NewSet(items...)
	req.True(set.Equal(NewSet(items...)), "same quad")
	req.False(set.Equal(NewSet(items[1:]...)), "so empty without the first")
	req.False(set.Equal(NewSet(append([]string{"1a"}, items[1:]...)...)), "swapped first item")

	counter := 0
	set.ForEach(func(_ string, stop func()) {
		counter++
		stop()
	})
	req.Equal(1, counter, "immediately stopped")

	set.ForEach(func(s string, _ func()) {
		found := -1
		for i, each := range items {
			if each == s {
				found = i
				break
			}
		}
		req.NotEqual(-1, found)
		diminished := append(items[0:found], items[found+1:]...)
		items = diminished
	})
	req.Empty(items, "ForEach should've removed all items")

	set = NewSet("m1", "o2", "o2", "n3")
	sliced := set.ToSlice()
	sort.Strings(sliced)
	req.Equal([]string{"m1", "n3", "o2"}, sliced)
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

func TestFilter(t *testing.T) {
	require.Equal(
		t,
		[]int{1, 2, 4},
		Filter(
			[]int{1, 2, 3, 4},
			func(i int) bool { return i%3 != 0 }))
	require.Nil(t, Filter([]int{1, 2, 3}, func(_ int) bool { return false }))
}

func TestOrPanic2(t *testing.T) {
	req := require.New(t)
	req.Equal("wow", OrPanic2("wow", nil)(""))
	req.PanicsWithValue(
		"Message: killed. Error: turn.",
		func() { OrPanic2("", fmt.Errorf("turn"))("killed") })
}
