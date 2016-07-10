package main

import (
	"fmt"
	"log"
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
	UniqueMatches  map[string]int
	BaseElo        int
	PendingMatches []*PendingMatch
	matchMaker     func() *PendingMatch
	wg             sync.WaitGroup
	StartTime      time.Time
	EndTime        time.Time

	playerUpdate        chan *playerRequest
	playerAdd           chan *playerRequest
	getPlayer           chan *playerRequest
	updateMatch         chan *Match
	updateUniqueMatches chan string
	db                  *bolt.DB
	dataBuckets         []string
	dbname              string
	TotalPlayers        int
}

type playerRequest struct {
	ID  uint64
	p   *Player
	ret chan *Player
	err error
}

type matchResult struct {
	m    *Match
	done chan struct{}
}

// Create a new sim and set the base elo with be
func NewEloSim(be int, dbname string) *EloSim {
	return &EloSim{UniqueMatches: make(map[string]int),
		playerUpdate:        make(chan *playerRequest, 1000000),
		playerAdd:           make(chan *playerRequest, 1000000),
		getPlayer:           make(chan *playerRequest, 1000000),
		updateMatch:         make(chan *Match, 1000000),
		updateUniqueMatches: make(chan string, 1000000),
		BaseElo:             be,
		dbname:              dbname,
		dataBuckets:         simDataBuckets}
}

// Start set elo sim background tasks
func (es *EloSim) Start() {
	es.StartTime = time.Now()

	// open database
	var err error
	es.db, err = bolt.Open(es.dbname, 0600, nil)
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
			case msg := <-es.playerAdd:
				es.db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(players))
					mb, err := msg.p.GobEncode()
					if err != nil {
						println("player gob fail")
						return err
					}
					err = b.Put([]byte(strconv.FormatUint(msg.ID, 16)), mb)

					return nil
				})
				es.TotalPlayers = es.TotalPlayers + 1

				msg.ret <- msg.p
				es.wg.Done()
			case msg := <-es.playerUpdate:
				es.db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(players))
					mb, err := msg.p.GobEncode()
					if err != nil {
						println("player gob fail")
						return err
					}
					err = b.Put([]byte(strconv.FormatUint(msg.ID, 16)), mb)
					bs := b.Stats()
					es.TotalPlayers = bs.KeyN
					return nil
				})
				msg.ret <- msg.p
				es.wg.Done()
			case msg := <-es.getPlayer:
				es.db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(players))
					v := b.Get([]byte(strconv.FormatUint(msg.ID, 16)))
					p := &Player{}
					err := p.GobDecode(v)
					if err != nil {
						return err
					}
					msg.ret <- p
					return nil
				})
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

// AddPlayer add an eligible player to the simulation
func (es *EloSim) AddPlayer() uint64 {
	newID := es.newPlayerID()
	p := &Player{}
	p.CreatedAt = time.Now()
	p.Elo = es.BaseElo
	p.ID = newID
	pr := &playerRequest{p: p, ret: make(chan *Player)}
	es.wg.Add(1)
	es.playerUpdate <- pr
	msg := <-pr.ret
	return msg.ID
}

func (es *EloSim) UpdatePlayer(p *Player) uint64 {
	es.wg.Add(1)
	pr := &playerRequest{p: p, ret: make(chan *Player)}
	es.playerUpdate <- pr
	msg := <-pr.ret
	return msg.ID
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
	return uint64(es.TotalPlayers)
}

// newMatchID generate a new match ID
func (es *EloSim) newMatchID() uint64 {
	rand.Seed(time.Now().UnixNano() * rand.Int63() / 2)
	dt := time.Now().UnixNano() * rand.Int63()

	return uint64(dt)
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

	log.Println("TP XXX", es.TotalPlayers)
	tp := es.TotalPlayers
	if tp < 2 {
		tp = 2
	}
	rand.Seed(time.Now().UnixNano() * rand.Int63() / 2)
	pa = uint64(rand.Int63n(int64(tp - 1)))

	rand.Seed(time.Now().UnixNano() * rand.Int63() * 2)
	pb = uint64(rand.Int63n(int64(tp - 1)))

	log.Println("Player chosesn")
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
		log.Println("got players")
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
	esr.TotalPlayers = es.TotalPlayers
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
	/*for _, v := range es.Players {
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

	}*/
	esr.AverageElo = esr.AverageElo / es.TotalPlayers

	// bracket players

	/*	for _, v := range es.Players {
			if v.Elo <= esr.AverageElo {
				esr.EloBrackets[0] = esr.EloBrackets[0] + 1
			} else if v.Elo > esr.AverageElo && v.Elo <= esr.AverageElo+((esr.HighestElo-esr.AverageElo)/3) {
				esr.EloBrackets[1] = esr.EloBrackets[1] + 1
			} else if v.Elo > esr.AverageElo+((esr.HighestElo-esr.AverageElo)/3) {
				esr.EloBrackets[2] = esr.EloBrackets[2] + 1
			}
		}
	*/
	return esr
}
