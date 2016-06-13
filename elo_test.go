package main

import "testing"

func TestCalcELO(t *testing.T) {
	p := 1000
	o := 1000

	elo := calcELO(p, o, true)
	t.Log(elo)

	elo = calcELO(o, p, false)
	t.Log(elo)
}
