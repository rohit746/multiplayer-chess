package main

import (
	"sync"

	"github.com/notnil/chess"
)

type Game struct {
	Game    *chess.Game
	Players []*Player
	sync.Mutex
}
