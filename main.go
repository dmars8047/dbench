package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

const logo = `██████╗ ██████╗ ███████╗███╗   ██╗ ██████╗██╗  ██╗
██╔══██╗██╔══██╗██╔════╝████╗  ██║██╔════╝██║  ██║
██║  ██║██████╔╝█████╗  ██╔██╗ ██║██║     ███████║
██║  ██║██╔══██╗██╔══╝  ██║╚██╗██║██║     ██╔══██║
██████╔╝██████╔╝███████╗██║ ╚████║╚██████╗██║  ██║
╚═════╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝
                                                  `

func main() {
	fmt.Printf("\n%s\n", logo)

	var p int
	var all bool

	flag.IntVar(&p, "proc", -1, "Number of logical processors. Not providing this will default to the number of logical processors available.")
	flag.BoolVar(&all, "all", false, "Indicates that the benchmark should be run on all logical processor (cores).")

	flag.Parse()

	// Run a single core benchmark by default
	var logicalProcessors int = 1

	if all {
		logicalProcessors = getLogicalProcessors()

		if logicalProcessors <= 0 {
			fmt.Printf("Logical processor count could not be parsed. Exiting.\n")
			return
		}

		fmt.Printf("Logical processors detected - %d\n\n", logicalProcessors)
	} else if p > 0 {
		logicalProcessors = p

		fmt.Printf("Logical processors - %d\n\n", logicalProcessors)
	} else {
		fmt.Printf("Running single core benchmark.\n\n")
	}

	ct := NewCycleTracker(logicalProcessors)

	ctx, cancel := context.WithCancel(context.Background())

	totalStartTime := time.Now()

	fmt.Printf("Running Prime Sieve Benchmark\n")

	for x := range logicalProcessors {
		go func(f int) {
			sw := ct.seiveBks[f]
			for {
				select {
				case <-ctx.Done():
					return
				default:
					sw.Execute()
				}
			}
		}(x)
	}

	time.Sleep(time.Second * 30)
	cancel()

	// fmt.Printf("Average Iteration Duration - %f seconds\n", ct.Average())
	fmt.Printf("\nOverall Duration - %f seconds\n", time.Since(totalStartTime).Seconds())
	fmt.Printf("Score - %d\n\n", ct.Score())
}

func getLogicalProcessors() int {
	var cores int = 0

	switch os := runtime.GOOS; os {
	case "darwin":
		// Run Command sysctl -n hw.logicalcpu
		cmd := exec.Command("sysctl", "-n", "hw.logicalcpu")

		output, err := cmd.Output()

		if err != nil {
			panic(err)
		}

		cmd.Start()

		if len(output) == 0 {
			return -1
		}

		// Check if the last character is a \n
		var nl byte = 0xa

		if output[len(output)-1] == nl {
			output = output[:len(output)-1]
		}

		cores, err = strconv.Atoi(string(output))

		if err != nil {
			return -1
		}
	case "linux":
		// Run Command nproc
		cmd := exec.Command("nproc")

		output, err := cmd.Output()

		if err != nil {
			panic(err)
		}

		cmd.Start()

		if len(output) == 0 {
			return -1
		}

		// Check if the last character is a \n
		var nl byte = 0xa

		if output[len(output)-1] == nl {
			output = output[:len(output)-1]
		}

		cores, err = strconv.Atoi(string(output))

		if err != nil {
			return -1
		}
	case "windows":
		return -1
	case "freebsd":
		return -1
	default:
		return -1
	}

	return cores
}

type CycleTracker struct {
	seiveBks []*SieveBenchmark
}

func NewCycleTracker(cores int) *CycleTracker {
	sbk := make([]*SieveBenchmark, 0)

	for range cores {
		sbk = append(sbk, NewSieveBenchmark())
	}

	return &CycleTracker{
		seiveBks: sbk,
	}
}

func (ct *CycleTracker) Score() uint64 {
	res := uint64(0)

	for _, s := range ct.seiveBks {
		res += s.cnt
	}

	f := float64(res) / 10000.0

	return uint64(math.Floor(f))
}

type SieveBenchmark struct {
	n    uint64
	nums []bool
	res  []uint64
	cnt  uint64
}

func NewSieveBenchmark() *SieveBenchmark {
	const n uint64 = 1000

	return &SieveBenchmark{
		n:    n,
		nums: make([]bool, n+1),
		res:  make([]uint64, 0, 168),
		cnt:  0,
	}
}

func (s *SieveBenchmark) Execute() {

	s.res = s.res[:0]

	for i := range s.nums[2:] {
		s.nums[i+2] = true
	}

	var p uint64

	for p = 2; p*p <= s.n; p++ {
		if s.nums[p] {
			for i := p * p; i <= s.n; i += p {
				s.nums[i] = false
			}
		}
	}

	var i uint64

	for i = 2; i <= s.n; i++ {
		if s.nums[i] {
			s.res = append(s.res, i)
		}
	}

	s.cnt++
}
