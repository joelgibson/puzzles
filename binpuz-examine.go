package main

import (
	"./binpuz"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const maxcount = 5

var soln = flag.Bool("solution", false, "Show solution")
var verb = flag.Bool("working", false, "Shows working out")
var diff = flag.Bool("difficulty", false, "Information on difficulty")

func main() {
	flag.Parse()

	puzz, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	puzzs := strings.TrimSpace(string(puzz))
	p, err := binpuz.FromString(puzzs)
	if err != nil {
		fmt.Println(err)
		return
	}

	nsolns := p.CountSolns(maxcount)
	fmt.Printf("Counted %d solutions (up to a maximum of %d)\n", nsolns, maxcount)
	if nsolns != 1 {
		return
	}
	if *soln {
		fmt.Println("Puzzle:")
		fmt.Println(p)
		fmt.Println("Solution:")
		solns := p.ListSolns()
		fmt.Println(solns[0])
	}

	s, steps, _ := p.Solve()
	if *verb {
		fmt.Printf("\n\nHow to solve:\n\n")
		fmt.Println(p)

		p := p.Clone()
		for _, step := range steps {
			p.Apply(step.Changes)
			fmt.Printf("\n%s (Difficulty %d)\n%v\n", step.Reason, step.Diff, p)
		}
	}

	if *diff {
		fmt.Printf("\n\nDifficulty details:\n\n")
		if !s.Solved() {
			fmt.Println("WARNING: Puzzle not fully solved")
		}
		sum := 0
		m := make(map[int]int)
		max := 0
		tot := 0
		for _, step := range steps {
			m[step.Diff]++
			if max < step.Diff {
				max = step.Diff
			}
			if step.Diff == 0 {
				continue
			}
			sum += step.Diff
			tot++
		}
		fmt.Printf("Solution length: %d (incl. 0's)\n", len(steps))
		fmt.Printf("Average difficulty: %f (excl. 0's)\n", float64(sum)/float64(tot))
		fmt.Printf("Maximum difficulty: %d\n", max)

		fmt.Printf("Difficulty breakdown:\n")
		for i := 0; i < 10; i++ {
			fmt.Printf("  Difficulty %d steps: %d\n", i, m[i])
		}
	}
}
