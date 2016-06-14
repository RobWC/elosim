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

func TestMegaEloContest(t *testing.T) {

	players := 1000
	playerMap := make(map[int]int)
	matchMap := make(map[int]int)
	combo := make(map[string]int)
	var eloRank ByElo

	resultChan := make(chan testEloResult)

	for i := 0; i > players; i++ {
		playerMap[i] = 1050
		matchMap[i] = 0
	}

	round := 0
	maxRounds := 20 * players
	var wg sync.WaitGroup

	go func() {
		for msg := range resultChan {
			playerMap[msg.p] = msg.pelo
			playerMap[msg.o] = msg.oelo

			matchMap[msg.p] = matchMap[msg.p] + 1
			matchMap[msg.o] = matchMap[msg.o] + 1
		}
	}()

	for {

		r := testEloResult{}
		rand.Seed(time.Now().UnixNano())
		r.p = rand.Intn(players)
		time.Sleep(4 * time.Nanosecond)
		rand.Seed(time.Now().UnixNano() / 2)
		r.o = rand.Intn(players)

		combo[strconv.Itoa(r.p)+strconv.Itoa(r.o)] = combo[strconv.Itoa(r.p)+strconv.Itoa(r.o)] + 1

		if r.p == r.o {
			continue
		}

		wg.Add(1)
		go func() {

			defer wg.Done()
			c := rand.Float64()
			win := false
			if c > calcEloWinChance(r.p, r.o) {
				win = true
			}
			r.pelo, r.oelo = eloBattle(playerMap[r.p], playerMap[r.o], win)

			resultChan <- r
		}()

		if round == maxRounds {
			break
		}

		round = round + 1
	}

	wg.Wait()

	for _, v := range playerMap {
		eloRank = append(eloRank, v)
	}

	sort.Sort(ByElo(eloRank))
	t.Log(eloRank)

	calcTotal := 0
	for _, v := range matchMap {
		calcTotal = calcTotal + v
	}
	t.Logf("Total Matches %d Total Played %d Last Round %d Total Unique %d\n", maxRounds, calcTotal/2, round, len(combo))

	// bucket groups
	eloBucket := make(map[string]int)

	eloBucket["bronze"] = 0
	eloBucket["silver"] = 0
	eloBucket["gold"] = 0

	for _, v := range playerMap {
		if v <= eloRank[len(eloRank)-1]/3 {
			eloBucket["bronze"] = eloBucket["bronze"] + 1
		} else if v < eloRank[len(eloRank)-1]/2+1 {
			eloBucket["silver"] = eloBucket["siver"] + 1
		} else {
			eloBucket["gold"] = eloBucket["gold"] + 1
		}
	}

	for k, v := range eloBucket {
		t.Logf("Bucket %s Total %d", k, v)
	}
	t.Log("Total Players", players)

}
