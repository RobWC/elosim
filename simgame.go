package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const (
	BasePlayerID = 90000
	IDIncrement  = 32
	matchHistory = "matchHistory"
	players      = "players"
	uniMatches   = "uniqueMatches"
)

var simDataBuckets = []string{matchHistory, players, uniMatches}

// EloSim an elo simulation system
type EloSim struct {
	Players        map[uint64]*Player
	UniqueMatches  map[string]int
	BaseElo        int
	PendingMatches []*PendingMatch
	matchMaker     func() *PendingMatch
	wg             sync.WaitGroup
	StartTime      time.Time
	EndTime        time.Time

	playerUpdate        chan *Player
	getPlayer           chan *playerRequest
	updateMatch         chan *Match
	updateUniqueMatches chan string
	db                  *bolt.DB
	dataBuckets         []string
}

type playerRequest struct {
	ID  uint64
	ret chan *Player
}

type matchResult struct {
	m    *Match
	done chan struct{}
}

// Create a new sim and set the base elo with be
func NewEloSim(be int) *EloSim {
	return &EloSim{Players: make(map[uint64]*Player),
		UniqueMatches:       make(map[string]int),
		playerUpdate:        make(chan *Player, 1000000),
		getPlayer:           make(chan *playerRequest, 1000000),
		updateMatch:         make(chan *Match, 1000000),
		updateUniqueMatches: make(chan string, 1000000),
		BaseElo:             be,
		dataBuckets:         simDataBuckets}
}

// Start set elo sim background tasks
func (es *EloSim) Start() {
	es.StartTime = time.Now()

	// open database
	var err error
	es.db, err = bolt.Open("sim.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	es.db.Update(func(tx *bolt.Tx) error {
		for _, v := range es.dataBuckets {
			_, err := tx.CreateBucketIfNotExists([]byte(v))
			if err != nil {
				return fmt.Errorf("unable to create bucket: %s", err)
			}
		}
		return nil
	})

	go func() {
		for {
			select {
			case msg := <-es.playerUpdate:

				es.Players[msg.ID] = msg
				es.wg.Done()
			case msg := <-es.getPlayer:
				msg.ret <- es.Players[msg.ID]
				es.wg.Done()
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-es.updateMatch:
				es.db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(matchHistory))
					mb, err := msg.GobEncode()
					if err != nil {
						println("gob fail")
						return err
					}
					err = b.Put([]byte(strconv.FormatUint(msg.ID, 16)), mb)
					return nil
				})
				es.wg.Done()
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-es.updateUniqueMatches:
				es.UniqueMatches[msg] = es.UniqueMatches[msg] + 1
				es.wg.Done()
			}
		}
	}()

}

func (es *EloSim) GetPlayer(id uint64) *Player {
	es.wg.Add(1)
	pr := &playerRequest{ID: id, ret: make(chan *Player)}
	es.getPlayer <- pr
	return <-pr.ret
}

func (es *EloSim) UpdatePlayer(p *Player) {
	es.wg.Add(1)
	es.playerUpdate <- p
}

func (es *EloSim) UpdateUniqueMatch(m string) {
	es.wg.Add(1)
	es.updateUniqueMatches <- m
}

func (es *EloSim) Stop() {
	es.wg.Wait()
	es.EndTime = time.Now()
}

// newPlayerID gernerate a new player ID
func (es *EloSim) newPlayerID() uint64 {
	return uint64(len(es.Players))
}

// newMatchID generate a new match ID
func (es *EloSim) newMatchID() uint64 {
	rand.Seed(time.Now().UnixNano() * rand.Int63() / 2)
	dt := time.Now().UnixNano() * rand.Int63()

	return uint64(dt)
}

// AddPlayer add an eligible player to the simulation
func (es *EloSim) AddPlayer() uint64 {
	newID := es.newPlayerID()
	p := &Player{}
	p.CreatedAt = time.Now()
	p.Elo = es.BaseElo
	p.ID = newID
	if _, ok := es.Players[newID]; !ok {
		es.Players[newID] = p
	} else {
		es.AddPlayer()
	}
	return newID
}

func (es *EloSim) AddMatch(m *Match) {
	es.wg.Add(1)
	es.updateMatch <- m
}

// SetMatchMaking set the match making method to generate a PendingMatch
func (es *EloSim) SetMatchMaking(mm func() *PendingMatch) {
	es.matchMaker = mm
}

// GenerateMatch crate a new PendingMatch within the simulation
func (es *EloSim) GenerateMatch() *PendingMatch {
	return es.matchMaker()
}

func (es *EloSim) RandomSelectPlayers() *PendingMatch {
	pa, pb := uint64(0), uint64(0)

	rand.Seed(time.Now().UnixNano() * rand.Int63() / 2)
	pa = uint64(rand.Int63n(int64(len(es.Players) - 1)))

	rand.Seed(time.Now().UnixNano() * rand.Int63() * 2)
	pb = uint64(rand.Int63n(int64(len(es.Players) - 1)))

	return &PendingMatch{TeamA: []uint64{pa}, TeamB: []uint64{pb}}
}

func (es *EloSim) SimMatch(pm *PendingMatch) {
	es.wg.Add(1)
	match := &Match{ID: es.newMatchID(), TeamA: pm.TeamA, TeamB: pm.TeamB}
	match.Start()

	if len(match.TeamA) == 1 && len(match.TeamB) == 1 {
		win := false

		pa := es.GetPlayer(match.TeamA[0])
		pb := es.GetPlayer(match.TeamB[0])

		pa.StartGame()
		pb.StartGame()
		// add unique match check
		es.UpdateUniqueMatch(fmt.Sprintf("%X%X", match.TeamA[0], match.TeamB[0]))

		if rand.Float64() > calcEloWinChance(pa.Elo, pb.Elo) {
			win = true
			match.TeamAWin()
			pa.Wins = pa.Wins + 1
			pb.Losses = pb.Losses + 1
		} else {
			match.TeamBWin()
			pb.Wins = pb.Wins + 1
			pb.Losses = pa.Losses + 1
		}
		pa.Elo, pb.Elo = eloBattle(pa.Elo, pb.Elo, win)
		// add match history
		es.UpdatePlayer(pa)
		es.UpdatePlayer(pb)
		pa.EndGame()
		pb.EndGame()

	}

	match.Stop()
	es.AddMatch(match)
	es.wg.Done()
}

type EloSimReport struct {
	StartTime        time.Time
	EndTime          time.Time
	TotalPlayers     int
	UniqueMatches    int
	TotalMatches     int
	HighestElo       int
	HighestEloPlayer uint64
	LowestElo        int
	LowestEloPlayer  uint64
	AverageElo       int
	EloBrackets      []int
}

func (es *EloSim) FinalReport() *EloSimReport {
	esr := &EloSimReport{EloBrackets: []int{0, 0, 0}}

	esr.StartTime = es.StartTime
	esr.EndTime = es.EndTime
	esr.TotalPlayers = len(es.Players)
	err := es.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(matchHistory))
		bs := b.Stats()
		esr.TotalMatches = bs.KeyN
		return nil
	})
	if err != nil {
		println(err)
	}
	esr.UniqueMatches = len(es.UniqueMatches)

	// determine highest and lowest elo
	for _, v := range es.Players {
		esr.AverageElo = esr.AverageElo + v.Elo

		// check highest elo
		if v.Elo > esr.HighestElo {
			esr.HighestElo = v.Elo
			esr.HighestEloPlayer = v.ID
		}

		if esr.LowestElo > v.Elo || esr.LowestElo == 0 {
			esr.LowestElo = v.Elo
			esr.LowestEloPlayer = v.ID
		}

	}
	esr.AverageElo = esr.AverageElo / len(es.Players)

	// bracket players

	for _, v := range es.Players {
		if v.Elo <= esr.AverageElo {
			esr.EloBrackets[0] = esr.EloBrackets[0] + 1
		} else if v.Elo > esr.AverageElo && v.Elo <= esr.AverageElo+((esr.HighestElo-esr.AverageElo)/3) {
			esr.EloBrackets[1] = esr.EloBrackets[1] + 1
		} else if v.Elo > esr.AverageElo+((esr.HighestElo-esr.AverageElo)/3) {
			esr.EloBrackets[2] = esr.EloBrackets[2] + 1
		}
	}

	return esr
}
