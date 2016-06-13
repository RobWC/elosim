package main

type Stats struct {
	Health int
	Mana   int

	// physical
	Strength int
	Body     int

	// agility
	Dexterity int
	Speed     int

	// thinking
	Intelligence int
	Knowledge    int

	generated bool
}
