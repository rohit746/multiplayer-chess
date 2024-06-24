package main

import (
	"sync"

	"github.com/notnil/chess"
)

type Game struct {
	sync.Mutex
	Game    *chess.Game
	Players []*Player
}
