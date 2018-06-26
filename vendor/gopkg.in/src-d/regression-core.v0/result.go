package regression

import (
	"fmt"
	"time"
)

// Result struct holds resource usage from a run.
type Result struct {
	Memory int64
	Wtime  time.Duration
	Stime  time.Duration
	Utime  time.Duration
}

// Comparison has the percentages of change between two runs.
type Comparison struct {
	Memory float64
	Wtime  float64
	Stime  float64
	Utime  float64
}

// Compare returns percentage difference between this and another run.
func (p *Result) Compare(q *Result) Comparison {
	return Comparison{
		Memory: Percent(p.Memory, q.Memory),
		Wtime:  Percent(int64(p.Wtime), int64(q.Wtime)),
		Stime:  Percent(int64(p.Stime), int64(q.Stime)),
		Utime:  Percent(int64(p.Utime), int64(q.Utime)),
	}
}

var CompareFormat = "%s: %v -> %v (%v), %v\n"

// ComparePrint does a result comparison, prints the result in human readable
// form and returns a bool if change is within allowance.
func (p *Result) ComparePrint(q *Result, allowance float64) bool {
	ok := true
	c := p.Compare(q)

	if c.Memory > allowance {
		ok = false
	}
	fmt.Printf(CompareFormat,
		"Memory",
		p.Memory,
		q.Memory,
		c.Memory,
		allowance > c.Memory,
	)

	if c.Wtime > allowance {
		ok = false
	}
	fmt.Printf(CompareFormat,
		"Wtime",
		p.Wtime,
		q.Wtime,
		c.Wtime,
		allowance > c.Wtime,
	)

	fmt.Printf(CompareFormat,
		"Stime",
		p.Stime,
		q.Stime,
		c.Stime,
		allowance > c.Stime,
	)

	fmt.Printf(CompareFormat,
		"Utime",
		p.Utime,
		q.Utime,
		c.Utime,
		allowance > c.Utime,
	)

	return ok
}

// Percent returns the percentage difference between to int64.
func Percent(a, b int64) float64 {
	diff := b - a
	return (float64(diff) / float64(a)) * 100
}
