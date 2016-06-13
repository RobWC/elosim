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

	round := 0
	for {
		// start round

		if false {
			break
		}
		round = round + 1
	}

	return bs
}
