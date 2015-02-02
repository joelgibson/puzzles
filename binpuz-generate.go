package main

import (
	"./binpuz"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"time"
)

type Coord struct{ i, j int }
type Coords []Coord

func (c Coords) Len() int           { return len(c) }
func (c Coords) Less(i, j int) bool { return c[i].i < c[j].i || (c[i].i == c[j].i && c[i].j < c[j].j) }
func (c Coords) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

type Boards []binpuz.Board

func (b Boards) Len() int           { return len(b) }
func (b Boards) Less(i, j int) bool { return b[i].Count() < b[j].Count() }
func (b Boards) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func Shuffle(elems sort.Interface, r *rand.Rand) {
	len := elems.Len()
	for i := len - 1; i >= 0; i-- {
		elems.Swap(i, r.Intn(i+1))
	}
}

func CoordsFor(size int) []Coord {
	coords := make([]Coord, size*size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			coords[i*size+j] = Coord{i, j}
		}
	}
	return coords
}

func RandByte(r *rand.Rand) byte {
	if rand.Intn(2) == 0 {
		return binpuz.Zero
	}
	return binpuz.One
}

// Rate the difficulty of a board. Panics if the board is inconsistent.
// Returns -1 for unsolved boards.
func Diff(b binpuz.Board) int {
	s, steps, err := b.Solve()
	if err != nil {
		panic(err)
	}
	if !s.Solved() {
		return -1
	}
	diff := 0
	for _, step := range steps {
		if diff < step.Diff {
			diff = step.Diff
		}
	}
	return diff
}

// genFull will generate puzzles which have a unique solution, and pass them back on a channel.
func genFull(r *rand.Rand, out chan<- binpuz.Board) {
	size := 10
	coords := CoordsFor(size)
	for {
		// Build a puzzle by placing things randomly. We might need to backtrack here.
		board := binpuz.New(size)
		Shuffle(Coords(coords), r)
		var f func(idx int) bool
		f = func(idx int) bool {
			i, j := coords[idx].i, coords[idx].j
			b := RandByte(r)
			board.Set(i, j, b)
			solns := board.CountSolns(2)
			if solns == 1 || solns >= 2 && f(idx+1) {
				return true
			}
			if b == binpuz.Zero {
				b = binpuz.One
			} else {
				b = binpuz.Zero
			}
			board.Set(i, j, b)
			if f(idx + 1) {
				return true
			}
			return false
		}
		f(0)
		out <- board
	}
}

const reductions = 10

// I have a feeling that removing numbers off a board with a unique solution is "strictly decreasing",
// in that if a number cannot be removed at an earlier step, that same number will not be able to be
// removed at a later step.
func reduce(r *rand.Rand, in <-chan binpuz.Board, out chan<- binpuz.Board) {
	var coords []Coord
	for {
		board := <-in
		m := make(map[int]binpuz.Board)
		for reds := 0; reds < reductions; reds++ {
			board := board.Clone()
			coords = coords[:0]
			for i := 0; i < board.Size; i++ {
				for j := 0; j < board.Size; j++ {
					if board.Get(i, j) != binpuz.Empty {
						coords = append(coords, Coord{i, j})
					}
				}
			}
			m[Diff(board)] = board.Clone()
			Shuffle(Coords(coords), r)
			for _, coord := range coords {
				i, j := coord.i, coord.j
				change := board.Set(i, j, binpuz.Empty)
				if !board.HasUniqueSoln() {
					board.Undo(change)
				} else {
					m[Diff(board)] = board.Clone()
				}
			}
		}
		for _, v := range m {
			out <- v
		}

	}
}

const keep = 5

func main() {
	seed := time.Now().UTC().UnixNano()
	getRand := func() *rand.Rand {
		seed++
		return rand.New(rand.NewSource(seed))
	}
	c := make(chan binpuz.Board, 10)
	d := make(chan binpuz.Board)
	for i := 0; i < 4; i++ {
		go reduce(getRand(), c, d)
		go genFull(getRand(), c)
	}

	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		for range sigs {
			fmt.Println("got sig")
			stop <- struct{}{}
		}
	}()

	m := make(map[int][]binpuz.Board)
	count := 0
	mod := 1
loop:
	for {
		select {
		case board := <-d:
			diff := Diff(board)
			m[diff] = append(m[diff], board)

			if len(m[diff]) > keep {
				sort.Sort(Boards(m[diff]))
				m[diff] = m[diff][:keep]
			}
			count++
			if count%mod == 0 {
				fmt.Println(count, "boards collected")
			}
			if count == mod*10 {
				mod *= 10
			}
		case _ = <-stop:
			break loop
		}
	}

	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		boards := m[k]
		sort.Sort(Boards(boards))
		fmt.Println("Difficulty", k, "boards:")
		for _, board := range boards {
			fmt.Printf("%v\n(%d numbers)\n\n", board, board.Count())
		}
		fmt.Println("-------")
	}
}
