package binpuz

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// The solver steps are each structured as a func(Board) Step, and each will output their
// "least difficult" step when asked. The solver then will run through and attempt to somehow
// prioritise easier steps over harder steps, and return a []Step of all of the steps they've taken.

// The solver steps do not mutate the board.

// flip takes a Zero or One byte and returns the opposite one.
func flip(b byte) byte {
	if b == Zero {
		return One
	}
	return Zero
}

// rowcol returns "row" if the board is not transposed, and "column" otherwise.
func (b Board) rowcol() string {
	if b.trans {
		return "column"
	}
	return "row"
}

var SolveErr = errors.New("Solving led to an inconsistent configuration")

// fixedRepls directly applies the no-three-adjacent rule, by making substitions like "00." => "001"
// and "0.0" => "010". It will keep applying these until it can make no progress, then bundle them
// all up and hand back a Step of difficulty 0.
func fixedRepls(b Board) Step {
	var changes []Change
	for progress := true; progress; {
		progress = false
		for _, q := range b.Views() {
			for i, row := range q.Rows {
				for j := 0; j < len(row)-2; j++ {
					a, b, c := row[j], row[j+1], row[j+2]
					idx, repl := -1, byte(0)

					// Pattern "00." and "11."
					if a != Empty && a == b && c == Empty {
						idx, repl = j+2, flip(a)
					}

					// Pattern ".00" and ".11"
					if a == Empty && b != Empty && b == c {
						idx, repl = j, flip(b)
					}

					// Pattern "0.0" and "1.1"
					if a != Empty && b == Empty && a == c {
						idx, repl = j+1, flip(a)
					}

					if idx >= 0 {
						changes = append(changes, q.Set(i, idx, repl))
						progress = true
					}
				}
			}
		}
	}
	b.Unapply(changes)

	return Step{
		Changes: changes,
		Diff:    0,
		Reason:  "Apply simple patterns",
	}
}

// remainingNos looks for rows or columns with all numbers of one type filled up.
// If a row or column has only one number missing, it gives a difficulty of 1. Otherwise,
// it gives a difficulty of 2.
func remainingNos(b Board) Step {
	cheap := Step{Diff: -1}
	for _, b := range b.Views() {
		for i, row := range b.Rows {
			if rowFull(row) {
				continue
			}

			var repl byte
			if zero, one := countZeroOne(row); zero == b.Size/2 {
				repl = One
			} else if one == b.Size/2 {
				repl = Zero
			} else {
				continue
			}

			var changes []Change
			for j, c := range row {
				if c == Empty {
					changes = append(changes, b.ChangeFor(i, j, repl))
				}
			}
			diff := 2
			if len(changes) == 1 {
				diff = 1
			}

			step := Step{
				Reason:  fmt.Sprintf("Only %c's remain in %s %d", repl, b.rowcol(), i+1),
				Changes: changes,
				Diff:    diff,
			}
			if cheap.Diff < 0 || step.Diff < cheap.Diff {
				cheap = step
			}
		}
	}
	return cheap
}

// rowFull returns true if there are no empty cells in the row.
func rowFull(row []byte) bool {
	return bytes.IndexByte(row, Empty) < 0
}

// Used to limit when to try every possible completion of a row.
func ncr(n, r int) int {
	if n-r < r {
		r = n - r
	}
	prod := 1
	for i := n; i > n-r; i-- {
		prod *= i
	}
	for i := 2; i <= r; i++ {
		prod /= i
	}
	return prod
}

type smallChange struct {
	idx int
	b   byte
}

// completeRow returns a collection of change lists, which represent all valid
// completions (filling in the blanks with Zero or One) of the given row.
// It will return nil for an already complete row. It may also give up if it has to
// investigate too many choices.
func completeRow(b Board, rowidx int, fullValidate bool) [][]smallChange {
	row := b.Rows[rowidx]
	if rowFull(row) {
		return nil
	}

	solns := make([][]smallChange, 0)
	work := make([]smallChange, 0, b.Size)
	valid := b.Validate
	if !fullValidate {
		valid = b.smallValidate
	}
	zeros, ones := countZeroOne(row)

	maxChoices := 20
	choices := ncr(len(row)-zeros-ones, len(row)/2-zeros)
	if choices > maxChoices {
		return nil
	}

	// We have (# empty cells) to fill with zeros and ones, which is

	var f func(i int)
	f = func(i int) {
		for ; i < len(row) && row[i] != Empty; i++ {
		}
		if i == len(row) {
			// Testing shows it's better to defer this test to here, rather
			// than early exiting from f()
			if valid() {
				solns = append(solns, append([]smallChange(nil), work...))
			}
			return
		}
		// Testing triples before placement gives almost no benefit
		if zeros < b.Size/2 {
			b.Set(rowidx, i, Zero)
			work = append(work, smallChange{i, Zero})
			zeros++
			f(i + 1)
			work = work[:len(work)-1]
			b.Set(rowidx, i, Empty)
			zeros--
		}
		if ones < b.Size/2 {
			b.Set(rowidx, i, One)
			work = append(work, smallChange{i, One})
			ones++
			f(i + 1)
			work = work[:len(work)-1]
			b.Set(rowidx, i, Empty)
			ones--
		}
		return
	}
	f(0)
	return solns
}

// intMin returns min(a, b)
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// completeRows will, for every row and column individually, find all possible completions
// of them. If for a row/col there is only one completion, that one is taken and given a difficulty
// of 3. If there are multiple valid completions, but one or more cells are constant over those
// completions, that is then taken and given a difficulty of 3 + min(# unused zeros, # unused ones).
// All of these difficulties are increased by one if the equal rows/cols constraint had to be used when
// finding them.
func completeRows(b Board, fullValidate bool) Step {
	var steps []Step
	baseDiff := 3
	if fullValidate {
		baseDiff++
	}
	for _, b := range b.Views() {
		for rowidx, row := range b.Rows {
			solns := completeRow(b, rowidx, fullValidate)
			if len(solns) == 0 {
				continue
			}
			var changes []Change

			// Was there only one way of completing the row?
			if len(solns) == 1 {
				for _, ch := range solns[0] {
					changes = append(changes, b.ChangeFor(rowidx, ch.idx, ch.b))
				}
				step := Step{
					Reason:  fmt.Sprintf("Only possible arrangement in %s %d", b.rowcol(), rowidx+1),
					Diff:    baseDiff,
					Changes: changes,
				}
				steps = append(steps, step)
				continue
			}

			// Now look for any common cells across all possible solutions. Here would usually use a
			// map[Change]int, but we want to avoid this for performance reasons.

			// A[0] <=> {0, '0'}, A[1] <=> {0, '1'}, A[2] <=> {1, '0'} and so on.
			allChanges := make([]int, 2*len(row))
			for _, soln := range solns {
				for _, ch := range soln {
					idx := ch.idx * 2
					if ch.b == One {
						idx++
					}
					allChanges[idx]++
				}
			}

			var inAll []Change
			for i, count := range allChanges {
				if count == len(solns) {
					repl := byte(Zero)
					if i%2 == 1 {
						repl = One
					}
					inAll = append(inAll, b.ChangeFor(rowidx, i/2, repl))
				}
			}

			// The hardness here is based on the fact that if, for example, 4/5 ones
			// are already placed in a row, it's easy to think about placing the last one, however
			// if not 3/5 ones and 3/5 zeros are placed, combinations are harder.
			if len(inAll) > 0 {
				frag := fmt.Sprintf("%d numbers are", len(inAll))
				if len(inAll) == 1 {
					frag = "1 number is"
				}
				reason := fmt.Sprintf("Out of %d possibilities in %s %d, %s common", len(solns), b.rowcol(), rowidx+1, frag)
				zeros, ones := countZeroOne(row)
				step := Step{
					Changes: inAll,
					Diff:    baseDiff + intMin(b.Size/2-zeros, b.Size/2-ones),
					Reason:  reason,
				}
				steps = append(steps, step)
				continue
			}
		}
	}
	sort.Sort(stepslice(steps))
	if len(steps) > 0 {
		return steps[0]
	}
	return Step{}
}

func completeRowsSmall(b Board) Step {
	return completeRows(b, false)
}
func completeRowsFull(b Board) Step {
	return completeRows(b, true)
}

var strats = []func(Board) Step{
	fixedRepls,
	remainingNos,
	completeRowsSmall,
	completeRowsFull,
}

// This is the general solver which will solve the board as far as possible using the given strategies.
// It returns a copy of the board which is as far as it got, the steps it used to get there, and possibly
// an error, reporting an inconsistency in the board.
func (b Board) Solve() (Board, []Step, error) {
	b = b.Clone()
	var steps []Step
	var err error
	for i := 0; i < len(strats); {
		step := strats[i](b)
		if len(step.Changes) > 0 {
			b.Apply(step.Changes)
			steps = append(steps, step)
			i = 0
		} else {
			i++
		}
		if !b.Validate() {
			err = SolveErr
			break
		}
	}
	return b, steps, err
}

// MaybeSolve is used to assist backtracking. It will mutate the board, but also return the changes
// needed to reverse it. If an inconsistency is caused, it will return false and back off it's changes.
func (b *Board) MaybeSolve() ([]Change, bool) {
	q, steps, err := b.Solve()
	if err != nil {
		return nil, false
	}

	var changes []Change
	for _, step := range steps {
		changes = append(changes, step.Changes...)
	}
	*b = q
	return changes, true
}

// Sometimes we need to choose between steps based on how difficult they are.
// This is the same as sorting the tuple (Diff, len(Changes))
type stepslice []Step

func (s stepslice) Len() int { return len(s) }
func (s stepslice) Less(i, j int) bool {
	return s[i].Diff < s[j].Diff || (s[i].Diff == s[j].Diff && len(s[i].Changes) < len(s[j].Changes))
}
func (s stepslice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
