package main

import "time"

type Player struct {
	ID        uint64
	CreatedAt time.Time
	Elo       int
	Wins      uint
	Losses    uint
}
