package binpuz

import (
	"bytes"
	"errors"
	"strings"
)

// Characters used for representing the board.
const (
	Empty = '.'
	Zero  = '0'
	One   = '1'
)

// A Board represents a partially complete puzzle.
type Board struct {
	// Size must be a multiple of 2
	Size int

	// Each of these are size*size slices, which must be
	// kept in sync (the easiest way is through the Set method)
	Rows, Cols [][]byte

	// Are we transposed? (For recording changes)
	trans bool
}

// A Change represents a change of one character on the board:
// replace 'a' at position (i, j) with 'b'
type Change struct {
	I, J int
	B    byte
}

// A step is a bundle of actions taken towards solving a puzzle,
// for a common reason.
type Step struct {
	// The changes making up the step
	Changes []Change

	// How difficult the step is
	Diff int

	// Argument explaining the step.
	Reason string
}

// Create a new blank board.
func New(size int) Board {
	if size <= 0 || size%2 != 0 {
		panic(errors.New("Size should be a positive even number!"))
	}

	// Allocate the board close together in memory, since we will
	// be working locally a lot.
	back := make([]byte, size*size*2)
	for i := range back {
		back[i] = Empty
	}
	rows := make([][]byte, size)
	for i := range rows {
		rows[i] = back[:size]
		back = back[size:]
	}
	cols := make([][]byte, size)
	for i := range cols {
		cols[i] = back[:size]
		back = back[size:]
	}
	return Board{size, rows, cols, false}
}

// Create a board from a string, formatted like "..\n01" or similar (any
// whitespace to separate lines will do).
func FromString(s string) (Board, error) {
	lines := bytes.Fields([]byte(strings.TrimSpace(s)))
	if len(lines) == 0 {
		return Board{}, errors.New("Board must contain data")
	}
	size := len(lines[0])
	if size%2 != 0 {
		return Board{}, errors.New("Board size must be even")
	}

	b := New(size)
	for i, line := range lines {
		if len(line) != size {
			return Board{}, errors.New("Inconsistent board size")
		}
		for j, c := range line {
			if !bytes.Contains([]byte{Zero, One, Empty}, []byte{c}) {
				return Board{}, errors.New("Invalid character in board")
			}
			b.Set(i, j, c)
		}
	}
	return b, nil
}

// Clone makes a copy of the board which shares no data with the original.
func (b Board) Clone() Board {
	q := New(b.Size)
	for i := 0; i < b.Size; i++ {
		for j := 0; j < b.Size; j++ {
			q.Set(i, j, b.Rows[i][j])
		}
	}
	return q
}

// Get returns the byte at position (i, j) on the board
func (b Board) Get(i, j int) byte { return b.Rows[i][j] }

// Set modifies the byte at position (i, j) on the board, and returns a Change
// such that Undo(change) will revert it. Set is aware of the board being
// transposed, and any changes returned by Set should be applied to an
// un-transposed board.
func (b Board) Set(i, j int, c byte) Change {
	b.Rows[i][j], b.Cols[j][i] = c, c
	return b.ChangeFor(i, j, c)
}

// ChangeFor returns the same thing as Set does, but does not mutate the board.
func (b Board) ChangeFor(i, j int, c byte) Change {
	if b.trans {
		return Change{j, i, c}
	}
	return Change{i, j, c}
}

// String returns the board as board.Size() lines.
func (b Board) String() string {
	var buf bytes.Buffer
	for i, row := range b.Rows {
		buf.Write(row)
		if i != b.Size-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// Undo reverts a single change.
func (b Board) Undo(change Change) {
	b.Set(change.I, change.J, Empty)
}

// Unapply reverts a collection of changes.
func (b Board) Unapply(changes []Change) {
	for _, change := range changes {
		b.Set(change.I, change.J, Empty)
	}
	return
}

// Apply applies a change to the board.
func (b Board) Apply(changes []Change) {
	for _, change := range changes {
		b.Set(change.I, change.J, change.B)
	}
	return
}

// Count returns how many cells in the board are nonempty.
func (b Board) Count() int {
	count := 0
	for _, row := range b.Rows {
		for _, c := range row {
			if c != Empty {
				count++
			}
		}
	}
	return count
}

// Views returns a [2]Board containing the current board, and the current
// board transposed.
func (b Board) Views() [2]Board {
	t := Board{Size: b.Size, Rows: b.Cols, Cols: b.Rows, trans: !b.trans}
	return [2]Board{b, t}
}
