package main

const (
	FigtherBaseHealth = 50
	FighterBaseMana   = 0
)

type Fighter struct {
	UnitID int
	ELO    int
	Name   string
	Stats  *Stats
}

func (f *Fighter) genStats() {
	s := &Stats{}
	s.Health = FigtherBaseHealth
	s.Mana = FighterBaseMana

	s.Strength = 7
	s.Body = 6

	s.Dexterity = 6
	s.Speed = 6

	s.Intelligence = 5
	s.Knowledge = 5

	s.generated = true
	f.Stats = s
}

func NewFighter(name string, unitid int) *Fighter {
	f := &Fighter{}
	f.UnitID = unitid
	f.Name = name
	f.ELO = 1000

	f.genStats()
	return f
}
