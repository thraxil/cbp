package main

type CircularLog struct {
	fails     []int64
	successes []int64
	size      int64
	head      int64

	// write channels
	advanceChan chan *advanceOp
	successChan chan *successOp
	failChan    chan *failOp

	// read channels
	totalChan   chan *totalOp
	percentChan chan *percentOp
}

func NewCircularLog(size int64) *CircularLog {
	// go always initializes to zero value
	fails := make([]int64, size)
	successes := make([]int64, size)
	c := CircularLog{
		fails:     fails,
		successes: successes,
		size:      size,
		head:      0,

		advanceChan: make(chan *advanceOp),
		successChan: make(chan *successOp),
		failChan:    make(chan *failOp),

		totalChan:   make(chan *totalOp),
		percentChan: make(chan *percentOp),
	}
	go c.run()
	return &c
}

func sum(nums []int64) (s int64) {
	for _, v := range nums {
		s += v
	}
	return s
}

type waitResp struct{}

type advanceOp struct{ Resp chan waitResp }
type successOp struct{ Resp chan waitResp }
type failOp struct{ Resp chan waitResp }

type readResp struct {
	T int64
	P float64
}

type totalOp struct{ Resp chan readResp }
type percentOp struct{ Resp chan readResp }

func (c *CircularLog) run() {
	for {
		select {
		// writes
		case op := <-c.advanceChan:
			c.advance()
			op.Resp <- waitResp{}
		case op := <-c.successChan:
			c.success()
			op.Resp <- waitResp{}
		case op := <-c.failChan:
			c.fail()
			op.Resp <- waitResp{}
			// reads
		case op := <-c.totalChan:
			v := c.total()
			op.Resp <- readResp{T: v}
		case op := <-c.percentChan:
			v := c.percent()
			op.Resp <- readResp{P: v}
		}
	}
}

func (c *CircularLog) sumFails() int64 {
	return sum(c.fails)
}

func (c *CircularLog) sumSuccesses() int64 {
	return sum(c.successes)
}

func (c *CircularLog) Total() int64 {
	r := make(chan readResp)
	c.totalChan <- &totalOp{r}
	return (<-r).T
}

func (c *CircularLog) total() int64 {
	return c.sumFails() + c.sumSuccesses()
}

func (c *CircularLog) Percent() float64 {
	r := make(chan readResp)
	c.percentChan <- &percentOp{r}
	return (<-r).P
}

func (c *CircularLog) percent() float64 {
	t := c.sumFails() + c.sumSuccesses()
	if t == 0 {
		return 0.0
	}
	return float64(c.sumFails()) / float64(t)
}

func (c *CircularLog) Advance() {
	wait := make(chan waitResp)
	c.advanceChan <- &advanceOp{wait}
	<-wait
}

func (c *CircularLog) advance() {
	c.head = (c.head + 1) % c.size
	c.successes[c.head] = 0
	c.fails[c.head] = 0
}

func (c *CircularLog) Success() {
	wait := make(chan waitResp)
	c.successChan <- &successOp{wait}
	<-wait
}

func (c *CircularLog) success() {
	c.successes[c.head] += 1
}

func (c *CircularLog) Fail() {
	wait := make(chan waitResp)
	c.failChan <- &failOp{wait}
	<-wait
}

func (c *CircularLog) fail() {
	c.fails[c.head] += 1
}
