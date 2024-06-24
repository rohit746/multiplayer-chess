package main

import (
	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
)

type Player struct {
	Conn  *websocket.Conn
	Color chess.Color
}
