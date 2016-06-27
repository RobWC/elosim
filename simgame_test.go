package main

import "testing"

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
	es.Start()
	playerCount := 10000
	for i := 0; i < playerCount; i++ {
		es.AddPlayer()
	}
	if len(es.Players) < playerCount {
		t.Fail()
	}

	t.Logf("Found %d players", playerCount)

	es.SetMatchMaking(es.RandomSelectPlayers)

	for i := 0; i < 100000; i++ {
		m := es.GenerateMatch()
		es.SimMatch(m)
	}
	es.Stop()
	//	for i := range es.MatchHistory {
	//		t.Logf("%#v", es.MatchHistory[i])
	//nid := fmt.Sprintf("%X%X", es.MatchHistory[i].TeamA[0], es.MatchHistory[i].TeamB[0])
	//t.Log(es.UniqueMatches[nid])
	//	}
	t.Logf("%#v", es.FinalReport())
}
