package binpuz

// HasSoln returns true if there exists any solution for the puzzle.
func (b Board) HasSoln() bool {
	b = b.Clone()
	b.Solve()
	return b.bruteHasSoln()
}
func (b Board) bruteHasSoln() bool {
	n, _ := b.bruteSolve(1, false)
	return n == 1
}

// HasUniqueSoln returns true if there exists exactly one solution for the
// puzzle.
func (b Board) HasUniqueSoln() bool {
	n, _ := b.bruteSolve(2, false)
	return n == 1
}

// CountSolns returns the number of distinct solutions for the puzzle, up to
// a maximum of limit. Setting limit to -1 counts as many as possible.
func (b Board) CountSolns(limit int) int {
	n, _ := b.bruteSolve(limit, false)
	return n
}

// ListSolns returns the distinct solutions for the puzzle.
func (b Board) ListSolns() []Board {
	_, solns := b.bruteSolve(0, true)
	return solns
}

// atMost limits the number of solutions returned (early exits). -1 to deactivate
// listSolns returns a list of the actual moves taken for every solution returned
func (b Board) bruteSolve(atMost int, listSolns bool) (nsolns int, solns []Board) {
	// Firstly, use the definite solution methods
	if soln, _, err := b.Solve(); err != nil {
		return 0, nil
	} else {
		b = soln
	}
	// Helper method: if it returns true, early exit pls
	var solve func(i, j int) bool
	solve = func(i, j int) bool {
		if !b.Validate() {
			return false
		}
		// Use our deterministic solver to accelerate this one: remember
		// to back out any changes it makes!
		changes, cont := b.MaybeSolve()
		defer func() { b.Unapply(changes) }()

		// If the deterministic solver ran into an inconsistency, this can't be solved.
		if !cont {
			return false
		}
		if j == b.Size {
			j = 0
			i += 1
		}
		for ; i < b.Size; j += 1 {
			if j == b.Size {
				j = 0
				i += 1
				if i == b.Size {
					break
				}
			}
			if b.Rows[i][j] == Empty {
				changes = append(changes, b.Set(i, j, Zero))
				if solve(i, j+1) {
					return true
				}
				b.Set(i, j, One)
				if solve(i, j+1) {
					return true
				}
				return false
			}
		}
		// If we're here, this is a valid solution
		nsolns++
		if listSolns {
			solns = append(solns, b.Clone())
		}
		if atMost >= 0 && nsolns >= atMost {
			return true
		}
		return false
	}
	solve(0, 0)
	return
}
