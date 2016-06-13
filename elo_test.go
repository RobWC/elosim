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

func TestExtendedELO(t *testing.T) {
	unita := 1000
	unitb := 1000
	unitc := 1200
	unitd := 800
	unite := 1900

	t.Log("A", unita, "B", unitb, "C", unitc, "D", unitd)

	// battle 1
	unita, unitb = eloBattle(unita, unitb, true)
	t.Log("A WINS", unita, unitb)

	// battle 2
	unita, unitc = eloBattle(unita, unitc, true)
	t.Log("A WINS", unita, unitc)

	// battle 3
	unita, unitd = eloBattle(unita, unitd, false)
	t.Log("A LOST", unita, unitd)

	// battle 4
	unita, unitb = eloBattle(unita, unitb, true)
	t.Log("A WINS", unita, unitb)

	// battle 5
	unita, unite = eloBattle(unita, unite, false)
	t.Log("A LOST", unita, unite)

	// battle 6
	unita, unitb = eloBattle(unita, unitb, false)
	t.Log("A LOST", unita, unitb)

	t.Log("A", unita, "B", unitb, "C", unitc, "D", unitd, "E", unite)
}
