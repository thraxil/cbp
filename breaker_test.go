package main

import "testing"

func TestBreakerCreate(t *testing.T) {
	NewBreaker(1.0, 0, nil)
}

func TestBreakerInitiallyNotTripped(t *testing.T) {
	b := NewBreaker(1.0, 0, nil)
	if b.Tripped() {
		t.Error("shouldn't be tripped yet")
	}
}

func TestSucessPassesThrough(t *testing.T) {
	c := NewCircularLog(3)
	b := NewBreaker(1.0, 0, c)
	b.Success()
	if c.Total() != 1 {
		t.Error("didn't register anything")
	}
	if c.Percent() != 0.0 {
		t.Error("shouldn't be any failures yet")
	}
	if b.Tripped() {
		t.Error("shouldn't be tripped yet")
	}
}

func TestFailPassesThrough(t *testing.T) {
	c := NewCircularLog(3)
	b := NewBreaker(1.0, 0, c)
	b.Fail()
	if c.Total() != 1 {
		t.Error("didn't register anything")
	}
	if c.Percent() != 1.0 {
		t.Error("should just be failure")
	}
	if !b.Tripped() {
		t.Error("should be tripped")
	}
}
