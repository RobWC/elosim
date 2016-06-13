package main

import "math"

const KFactor = 20

func calcELO(player, opponent int, win bool) int {
	w := 0

	if win {
		w = 1
	}

	f := math.Pow(10, float64(opponent-player)/400)

	return player + int(KFactor*(float64(w)-(f)))
}
