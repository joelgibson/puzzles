package binpuz

import "testing"

func TestCountEmptySolns(t *testing.T) {
	tests := []struct {
		size, expect int
	}{
		{2, 2},
		{4, 72},
		{6, 4140},
	}
	for _, test := range tests {
		solns := New(test.size).CountSolns(-1)
		if test.expect != solns {
			t.Errorf("Counted %d solns instead of %d for %d x %d board", solns, test.expect, test.size, test.size)
		}
	}
}
