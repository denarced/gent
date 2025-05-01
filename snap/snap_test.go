package snap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
