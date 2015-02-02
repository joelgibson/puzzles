package binpuz

import "bytes"

// countZeroOne returns the number of zeros and the number of ones
// in a byte slice.
func countZeroOne(s []byte) (int, int) {
	z, o := 0, 0
	for _, c := range s {
		if c == Zero {
			z++
		} else if c == One {
			o++
		}
	}
	return z, o
}

// checkThreeAdj returns true if the slice contains three consecutive equal
// bytes which are not Empty.
func checkThreeAdj(row []byte) bool {
	for j := 0; j < len(row)-2; j++ {
		a, b, c := row[j], row[j+1], row[j+2]
		if a != Empty && a == b && b == c {
			return true
		}
	}
	return false
}

// Validate returns true if the board obeys all the puzzle constraints,
// ignoring any Empty cells.
func (b Board) Validate() bool {
	for _, grid := range [][][]byte{b.Rows, b.Cols} {
		// Leave the capacity as a constant here to generate no garbage.
		fulls := make([][]byte, 0, 12)
		for _, row := range grid {
			// Three equal adjacent numbers
			if checkThreeAdj(row) {
				return false
			}

			// Numbers of ones and zeros
			zeros, ones := countZeroOne(row)
			if zeros > b.Size/2 || ones > b.Size/2 {
				return false
			}

			// Uniqueness of rows (ignore rows which are not full)
			if zeros+ones != b.Size {
				continue
			}
			for _, full := range fulls {
				if bytes.Equal(row, full) {
					return false
				}
			}
			fulls = append(fulls, row)
		}
	}
	return true
}

// Solved returns true if the board is both valid and full (no Empty characters).
func (b Board) Solved() bool {
	if !b.Validate() {
		return false
	}
	for _, row := range b.Rows {
		for _, c := range row {
			if c == Empty {
				return false
			}
		}
	}
	return true
}

// smallValidate ignores the equal rows/columns constraint. This is for use in
// difficulty grading.
func (b Board) smallValidate() bool {
	for _, grid := range [][][]byte{b.Rows, b.Cols} {
		for _, row := range grid {
			// Three equal adjacent numbers
			if checkThreeAdj(row) {
				return false
			}

			// Numbers of ones and zeros
			zeros, ones := countZeroOne(row)
			if zeros > b.Size/2 || ones > b.Size/2 {
				return false
			}
		}
	}
	return true
}
