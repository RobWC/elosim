package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	BasePlayerID = 90000
	IDIncrement  = 32
	matchHistory = "matchHistory"
	players      = "player"s
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
	getPlayer           chan *playerRequest
	updateMatch         chan *Match
	updateUniqueMatches chan string
	db                  *gorm.DB
	dataBuckets         []string
	dbname              string
	TotalPlayers        int
}

type playerRequest struct {
	p   *Player
	ret chan *Player
	err error
}

func newPlayerRequest(p *Player) *playerRequest {
	return &playerRequest{p: p, ret: make(chan *Player)}
}

type matchRequest struct {
	m   *Match
	ret chan *Match
	err error
}

func newMatchRequest(m *Match) *matchRequest {
	return &matchRequest{m: m, ret: make(chan *Match)}
}

// Create a new sim and set the base elo with be
func NewEloSim(be int, dbname string) *EloSim {
	return &EloSim{UniqueMatches: make(map[string]int),
		playerUpdate:        make(chan *playerRequest, 1000000),
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
	es.db, err = gorm.Open("sqlite3", es.dbname)
	if err != nil {
		panic(err)
	}

	// automigrate
	es.db.AutoMigrate(&Player{},&Match{})

	// player mangement
	go func() {
		for {
			select {
 			case msg := <-es.playerUpdate:
				if es.db.NewRecord(msg.p) {
					r, err := es.db.Create(&msg.p)
					msg.err = r.Error
					msg.p = r.Value
					es.TotalPlayers = es.TotalPlayers + 1
				} else {
					r := es.db.Save(msg.p)
					msg.err = e.Error
					msg.p = r.Value 
				}
				msg.ret <- msg.p
				es.wg.Done()
			case msg := <-es.getPlayer:
				r := es.db.First(&msg.p)
				msg.p = r.Value
				msg.err = r.Error
				msg.ret <- msg.p
				es.wg.Done()
			 
			}
		}
	}()

	// match management
	go func() {
		for {
			select {
			case msg := <-es.updateMatch:
				if es.db.NewRecord(msg.m) {
					r := es.db.Create(&msg)
					msg.p = r.Value
					msg.err = r.Error
				} else {
					// player already exists
					msg.err = fmt.Error("Match already exists")
				}
				msg.ret <- msg.p
				es.wg.Done()

			}
		}
	}()

}

func (es *EloSim) GetPlayer(id uint64) *Player {
	es.wg.Add(1)
	pr := &playerRequest{p: &Player{ID:id}, ret: make(chan *Player)}
	es.getPlayer <- pr
	return <-pr.ret
}

// AddPlayer add an eligible player to the simulation
func (es *EloSim) AddPlayer() uint64 {
	es.wg.Add(1)
 	p := &Player{}
 	p.Elo = es.BaseElo
	pr := &playerRequest{p: p, ret: make(chan *Player)}
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


func (es *EloSim) Stop() {
	es.wg.Wait()
	es.EndTime = time.Now()
}

func (es *EloSim) AddMatch(m *Match) *Match {
	es.wg.Add(1)
	mr := newMatchRequest(m)
	es.updateMatch <- mr
	return < mr.ret
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

func (es *EloSim) SimMatch(pm *PendingMatch) uint64 {
	es.wg.Add(1)
	match := &Match{TeamA: pm.TeamA, TeamB: pm.TeamB}
	match.Start()

	if len(match.TeamA) == 1 && len(match.TeamB) == 1 {
		win := false

		pa := es.GetPlayer(match.TeamA[0])
		pb := es.GetPlayer(match.TeamB[0])
		log.Println("got players")
		pa.StartGame()
		pb.StartGame()
		// add unique match check

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
	es.wg.Done()
	return es.AddMatch(match)
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
