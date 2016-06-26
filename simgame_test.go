package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestBasicEloEloSim(t *testing.T) {
	baseElo := 1200
	es := NewEloSim(baseElo)
	playerCount := 10000
	for i := 0; i < playerCount; i++ {
		es.AddPlayer()
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

func TestRandomMatchMakingEloSim(t *testing.T) {
	baseElo := 1200
	es := NewEloSim(baseElo)
	playerCount := 10000
	for i := 0; i < playerCount; i++ {
		es.AddPlayer()
	}
	if len(es.Players) < playerCount {
		t.Fail()
	}

	t.Logf("Found %d players", playerCount)

	es.SetMatchMaking(es.RandomSelectPlayers)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(10)) * time.Nanosecond)
			m := es.GenerateMatch()
			t.Logf("%d", m)
			es.SimMatch(m)
		}()
	}
	wg.Wait()
	t.Logf("Unique Matches %d Players %d", len(es.UniqueMatches), len(es.Players))
}
