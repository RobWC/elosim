package main

import "math"

const KFactor = 32

func calcELO(player, opponent int, win bool) int {
	w := 0

	if win {
		w = 1
	}

	f := 1 / (float64(1) + math.Pow(10, float64(opponent-player)/400))

	return player + int(KFactor*(float64(w)-(f)))
}

func eloBattle(player, opponent int, win bool) (int, int) {
	return calcELO(player, opponent, win), calcELO(opponent, player, !win)
}
