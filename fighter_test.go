package main

import "testing"

func TestNewFighter(t *testing.T) {
	name := "Dick Whitman"

	f := NewFighter(name, 1)

	if f.Name != name {
		t.Fatal()
	}

	t.Log(f)

}
