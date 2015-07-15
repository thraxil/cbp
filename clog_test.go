package main

import "testing"

func TestCreateCircularLog(t *testing.T) {
	NewCircularLog(10)
}

func TestInitialValues(t *testing.T) {
	c := NewCircularLog(10)
	if c.Percent() > 0.0 {
		t.Error("needs to start at 0.0. got", c.Percent())
	}
}
