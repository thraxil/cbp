package main

import (
	"math"
	"testing"
)

var EPSILON = 0.00001

// float not-equal comparison helper
func floatNE(a, b float64) bool {
	return math.Abs(a-b) > EPSILON
}

func TestCreateCircularLog(t *testing.T) {
	NewCircularLog(10)
}

func TestInitialValues(t *testing.T) {
	c := NewCircularLog(10)
	if c.Percent() > 0.0 {
		t.Error("needs to start at 0.0. got", c.Percent())
	}
}

func TestAdvance(t *testing.T) {
	c := NewCircularLog(1)
	c.Advance()
}

func TestSuccess(t *testing.T) {
	c := NewCircularLog(1)
	c.Success()
}

func TestFail(t *testing.T) {
	c := NewCircularLog(1)
	c.Fail()
	// 100% failure rate now
	if c.Percent() < 1.0 {
		t.Error("should be a 100%% failure rate")
	}
}

func TestRunThroughInPlace(t *testing.T) {
	c := NewCircularLog(1)
	if c.Percent() > 0.0 {
		t.Error("needs to start at 0.0. got", c.Percent())
	}
	c.Success()
	if c.Percent() > 0.0 {
		t.Error("expected 0.0. got", c.Percent())
	}
	c.Fail()
	if floatNE(c.Percent(), 0.5) {
		t.Error("expected 0.5. got", c.Percent())
	}
	c.Success()
	if floatNE(c.Percent(), 1./3.) {
		t.Error("expected 0.33. got", c.Percent())
	}
	c.Fail()
	if floatNE(c.Percent(), 0.5) {
		t.Error("expected 0.5. got", c.Percent())
	}
	c.Fail()
	if floatNE(c.Percent(), 3./5.) {
		t.Error("expected 0.5. got", c.Percent())
	}
}

func TestTotal(t *testing.T) {
	c := NewCircularLog(1)
	if c.Total() != 0 {
		t.Error("shouldn't be any in there yet")
	}
	c.Fail()
	if c.Total() != 1 {
		t.Error("expected 1, got", c.Total())
	}
	c.Success()
	if c.Total() != 2 {
		t.Error("expected 2, got", c.Total())
	}
}

func TestAdvanceWraps(t *testing.T) {
	c := NewCircularLog(3)
	// put a fail in the first slot
	c.Fail()
	// now, advancing the right number of steps
	// should reset it

	// starts at idx 0
	if floatNE(c.Percent(), 1.) {
		t.Error("not there yet")
	}

	// advance to 1
	c.Advance()
	if floatNE(c.Percent(), 1.) {
		t.Error("not there yet")
	}

	// advance to 2
	c.Advance()
	if floatNE(c.Percent(), 1.) {
		t.Error("not there yet")
	}

	// advance back to 1
	c.Advance()
	if floatNE(c.Percent(), 0.) {
		t.Error("should've reset")
	}

}
