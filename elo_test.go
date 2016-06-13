package main

import "testing"

func TestCalcELO(t *testing.T) {
	p := 1200
	o := 1100

	elo := calcELO(p, o, true)
	t.Log(elo)

	elo = calcELO(o, p, false)
	t.Log(elo)
}
