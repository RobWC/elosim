package main

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCalcELO(t *testing.T) {
	p := 1000
	o := 1000

	elo := calcELO(p, o, true)
	t.Log(elo)

	elo = calcELO(o, p, false)
	t.Log(elo)
}

func TestCalcELO2(t *testing.T) {
	p := 600
	o := 600

	elo := calcELO(p, o, true)
	t.Log(elo)

	elo = calcELO(o, p, false)
	t.Log(elo)
}

func TestExtendedELO(t *testing.T) {
	unita := 1000
	unitb := 1000
	unitc := 1200
	unitd := 800
	unite := 1900

	t.Log("A", unita, "B", unitb, "C", unitc, "D", unitd)

	// battle 1
	unita, unitb = eloBattle(unita, unitb, true)
	t.Log("A WINS", unita, unitb)

	// battle 2
	unita, unitc = eloBattle(unita, unitc, true)
	t.Log("A WINS", unita, unitc)

	// battle 3
	unita, unitd = eloBattle(unita, unitd, false)
	t.Log("A LOST", unita, unitd)

	// battle 4
	unita, unitb = eloBattle(unita, unitb, true)
	t.Log("A WINS", unita, unitb)

	// battle 5
	unita, unite = eloBattle(unita, unite, false)
	t.Log("A LOST", unita, unite)

	// battle 6
	unita, unitb = eloBattle(unita, unitb, false)
	t.Log("A LOST", unita, unitb)

	t.Log("A", unita, "B", unitb, "C", unitc, "D", unitd, "E", unite)
}

type ByElo []int

func (a ByElo) Len() int {
	return len(a)
}
func (a ByElo) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByElo) Less(i, j int) bool { return a[i] < a[j] }

type testEloResult struct {
	p    int
	o    int
	pelo int
	oelo int
}

type player struct {
	ID      int
	Wins    int
	Losses  int
	Matches int
	Elo     int
}

func TestMegaEloContest(t *testing.T) {

	players := 10000
	playerMap := make(map[int]*player)
	matchMap := make(map[int]int)
	combo := make(map[string]int)
	var eloRank ByElo

	resultChan := make(chan testEloResult)

	for i := 0; i < players; i++ {
		playerMap[i] = &player{ID: i, Elo: 1050}
		matchMap[i] = 0
	}

	round := 0
	maxRounds := 100 * players
	var wg sync.WaitGroup

	go func() {
		for msg := range resultChan {
			playerMap[msg.p].Elo = msg.pelo
			playerMap[msg.o].Elo = msg.oelo

			playerMap[msg.p].Matches = playerMap[msg.p].Matches + 1
			playerMap[msg.o].Matches = playerMap[msg.o].Matches + 1
		}
	}()

	for {

		r := testEloResult{}
		rand.Seed(time.Now().UnixNano())
		r.p = rand.Intn(players)
		rand.Seed(time.Now().UnixNano() / rand.Int63())
		r.o = rand.Intn(players)

		combo[strconv.Itoa(r.p)+strconv.Itoa(r.o)] = combo[strconv.Itoa(r.p)+strconv.Itoa(r.o)] + 1

		if r.p == r.o {
			continue
		}

		wg.Add(1)
		go func() {

			defer wg.Done()
			win := false
			if rand.Float64() > calcEloWinChance(r.p, r.o) {
				win = true
				playerMap[r.p].Wins = playerMap[r.p].Wins + 1
				playerMap[r.o].Losses = playerMap[r.o].Losses + 1
			} else {
				playerMap[r.o].Wins = playerMap[r.o].Wins + 1
				playerMap[r.p].Losses = playerMap[r.p].Losses + 1
			}
			r.pelo, r.oelo = eloBattle(playerMap[r.p].Elo, playerMap[r.o].Elo, win)

			resultChan <- r
		}()

		if round == maxRounds {
			break
		}

		round = round + 1
	}

	wg.Wait()

	for _, v := range playerMap {
		eloRank = append(eloRank, v.Elo)
	}

	sort.Sort(ByElo(eloRank))

	calcTotal := 0
	for _, v := range playerMap {
		calcTotal = calcTotal + v.Matches
	}
	t.Logf("Total Matches %d Total Played %d Last Round %d Total Unique %d\n", maxRounds, calcTotal/2, round, len(combo))

	totalElo := 0
	maxElo := 0
	highestEloPlayer := 0
	minElo := 0
	for k, v := range playerMap {
		totalElo = totalElo + v.Elo

		if v.Elo > maxElo {
			maxElo = v.Elo
			highestEloPlayer = k
		}

		if v.Elo < minElo || minElo == 0 {
			minElo = v.Elo
		}
	}

	avgElo := totalElo / len(playerMap)

	t.Logf("Average Elo %d Min Elo %d Max Elo %d", avgElo, minElo, maxElo)
	t.Logf("Highest Elo player %#v", playerMap[highestEloPlayer])
	// bucket groups
	eloBucket := make(map[string]int)

	eloBucket["bronze"] = 0
	eloBucket["silver"] = 0
	eloBucket["gold"] = 0

	t.Log("Total Players", players)

}
