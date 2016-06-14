package main

import "math"

const KFactorBase = 32
const MaxElo = 2400
const MinElo = 600

func calcEloWinChance(player, opponent int) float64 {
	return 1 / (float64(1) + math.Pow(10, float64((opponent-player)/400)))
}

func calcELO(player, opponent int, win bool) int {
	w := 0

	if win {
		w = 1
	}

	f := 1 / (float64(1) + math.Pow(10, float64((opponent-player)/400)))

	KFactor := KFactorBase
	if player >= 1200 && player <= 2000 {
		KFactor = KFactorBase - 8
	} else if player > 2000 {
		KFactor = KFactorBase - 16
	}

	e := player + int(float64(KFactor)*(float64(w)-(f)))

	if e > MaxElo {
		e = MaxElo
	} else if e < MinElo {
		e = MinElo
	}

	return e
}

func eloBattle(player, opponent int, win bool) (int, int) {
	return calcELO(player, opponent, win), calcELO(opponent, player, !win)
}
