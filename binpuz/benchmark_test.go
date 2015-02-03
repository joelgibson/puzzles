package binpuz

// go test -c && ./binpuz.test -test.cpuprofile=cpu.out -test.bench=. && go tool pprof binpuz.test cpu.out

import "testing"

// Common case, don't want this getting too slow.
func BenchmarkUniqueEmpty10x10(b *testing.B) {
	p := New(10)
	for i := 0; i < b.N; i++ {
		p.HasUniqueSoln()
	}
}
func BenchmarkUniqueEmpty12x12(b *testing.B) {
	p := New(12)
	for i := 0; i < b.N; i++ {
		p.HasUniqueSoln()
	}
}

func BenchmarkHasSolnEmpty14x14(b *testing.B) {
	p := New(14)
	for i := 0; i < b.N; i++ {
		p.HasSoln()
	}
}

func BenchmarkCountAll6x6(b *testing.B) {
	p := New(6)
	for i := 0; i < b.N; i++ {
		p.CountSolns(-1)
	}
}

func BenchmarkCountDifficult12x12(b *testing.B) {
	vhard := `.00....1..01
............
....0.0...0.
..1.00...1..
0..1....0...
0.........1.
.1..0.......
11..........
.......11...
.0.00.....1.
....0......1
1.......1.0.`
	p, _ := FromString(vhard)
	for i := 0; i < b.N; i++ {
		p.CountSolns(-1)
	}
}

// Keep this benchmark here since we don't want to suddenly become really
// slow at trying to solve empty boards.
func BenchmarkSolveEmpty14x14(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(14).Solve()
	}
}
