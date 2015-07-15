package main

type CircularLog struct {
	fails     []int64
	successes []int64
	size      int64
	head      int64
}

func NewCircularLog(size int64) *CircularLog {
	// go always initializes to zero value
	fails := make([]int64, size)
	successes := make([]int64, size)
	return &CircularLog{
		fails:     fails,
		successes: successes,
		size:      size,
		head:      0,
	}
}

func sum(nums []int64) (s int64) {
	for _, v := range nums {
		s += v
	}
	return s
}

func (c *CircularLog) sumFails() int64 {
	return sum(c.fails)
}

func (c *CircularLog) Percent() float64 {
	return 0.0
}
