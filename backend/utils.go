package main

import (
	"math/rand"

	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
)

func randomColor() chess.Color {
	if rand.Intn(2) == 0 {
		return chess.White
	}
	return chess.Black
}

func toggleColor(color chess.Color) chess.Color {
	if color == chess.White {
		return chess.Black
	}
	return chess.White
}

func getPlayerColor(ws *websocket.Conn, game *Game) chess.Color {
	for _, player := range game.Players {
		if player.Conn == ws {
			return player.Color
		}
	}
	return chess.NoColor
}
