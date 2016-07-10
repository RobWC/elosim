package main

import (
	"testing"
	"time"
)

func TestBasicEloEloSim(t *testing.T) {
	baseElo := 1200
	es := NewEloSim(baseElo, "tbees.db")
	es.Start()
	playerCount := 100
	for i := 0; i < playerCount; i++ {
		es.AddPlayer()
	}
	if es.TotalPlayers < playerCount {
		t.Fail()
	}

	t.Logf("Found %d players", playerCount)

	es.Stop()
}

func TestRandomMatchMakingEloSim(t *testing.T) {
	baseElo := 1050
	es := NewEloSim(baseElo, "trmmes.db")
	es.Start()
	playerCount := 1000
	for i := 0; i < playerCount; i++ {
		go es.AddPlayer()
	}
	if es.TotalPlayers < playerCount {
		t.Fail()
	}

	t.Logf("Found %d players", playerCount)

	es.SetMatchMaking(es.RandomSelectPlayers)
	for i := 0; i < 1000; i++ {
		time.Sleep(1 * time.Nanosecond)
		es.SimMatch(es.GenerateMatch())
	}

	es.Stop()
	//	for i := range es.MatchHistory {
	//		t.Logf("%#v", es.MatchHistory[i])
	//nid := fmt.Sprintf("%X%X", es.MatchHistory[i].TeamA[0], es.MatchHistory[i].TeamB[0])
	//t.Log(es.UniqueMatches[nid])
	//	}
	t.Logf("%#v", es.FinalReport())
}
