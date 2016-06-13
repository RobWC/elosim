package main

import "time"

type BattleStats struct {
	Winner int
	Loser  int
	Time   time.Time
}

func Battle(unita, unitb Unit) *BattleStats {
	bs := &BattleStats{}
	bs.Time = time.Now()

	// execute battle

	return bs
}
