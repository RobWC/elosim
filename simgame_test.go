package main

import "testing"

func TestBasicEloEloSim(t *testing.T) {
	baseElo := 1200
	es := NewEloSim(baseElo)
	playerCount := 10000
	for i := 0; i < playerCount; i++ {
		es.AddPlayer(&Player{})
	}
	if len(es.Players) < playerCount {
		t.Fail()
	}

	t.Logf("Found %d players", playerCount)

	checkLimit := 0
	for k, v := range es.Players {
		if v.Elo != baseElo {
			t.Fatal("Base Elo Missing")
		}
		pt, _ := v.CreatedAt.MarshalText()
		t.Logf("Player %X %X Created at %s", k, v.ID, pt)
		if checkLimit == 10 {
			break
		}
		checkLimit = checkLimit + 1
	}
}
