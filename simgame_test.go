package main

import (
	"log"
	"testing"
	"time"
)

func TestBasicEloEloSim(t *testing.T) {
	baseElo := 1200
	es := NewEloSim(baseElo, "tbees")
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
	es := NewEloSim(baseElo, "ttees")
	es.Start()

	log.Println("Making Players", time.Now())
	playerCount := 100
	for i := 0; i < playerCount; i++ {
		es.AddPlayer()
	}

	log.Println("Making Players Complete", time.Now())
	if es.TotalPlayers < playerCount {
		t.Fail()
	}

	log.Println("Making Matches", time.Now())
	es.SetMatchMaking(es.RandomSelectPlayers)
	for i := 0; i < 100000; i++ {
		es.SimMatch(es.GenerateMatch())
	}
	log.Println("Making Matches Complete", time.Now())

	es.Stop()
	t.Logf("%#v", es.FinalReport())
}
