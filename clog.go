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

func (c *CircularLog) sumSuccesses() int64 {
	return sum(c.successes)
}

func (c *CircularLog) Total() int64 {
	return c.sumFails() + c.sumSuccesses()
}

func (c *CircularLog) Percent() float64 {
	t := c.sumFails() + c.sumSuccesses()
	if t == 0 {
		return 0.0
	}
	return float64(c.sumFails()) / float64(t)
}

func (c *CircularLog) Advance() {
	c.head = (c.head + 1) % c.size
	c.successes[c.head] = 0
	c.fails[c.head] = 0
}

func (c *CircularLog) Success() {
	c.successes[c.head] += 1
}

func (c *CircularLog) Fail() {
	c.fails[c.head] += 1
}
