package main

import (
	"github.com/notnil/chess"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var games = make(map[string]*Game)
var gamesMutex sync.Mutex

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	log.Println("New connection established from", r.RemoteAddr)

	defer func() {
		log.Println("Connection closed from", r.RemoteAddr)
		removePlayer(ws)
		err := ws.Close()
		if err != nil {
			return
		}
	}()

	for {
		var msg map[string]string
		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Println("Unexpected close error:", err)
			} else {
				log.Println("Read error:", err)
			}
			break
		}
		log.Println("Received message:", msg)

		action := msg["action"]
		switch action {
		case "create":
			gameID := ksuid.New().String()
			playerColor := randomColor()
			player := &Player{Conn: ws, Color: playerColor}
			game := &Game{
				Game: chess.NewGame(),
				Players: []*Player{
					player,
				},
			}
			gamesMutex.Lock()
			games[gameID] = game
			gamesMutex.Unlock()
			err := ws.WriteJSON(map[string]string{"status": "created", "gameID": gameID, "color": playerColor.String()})
			if err != nil {
				return
			}
			log.Printf("Game created with ID: %s", gameID)
		case "join":
			gameID := msg["gameID"]
			gamesMutex.Lock()
			game, exists := games[gameID]
			if exists {
				if len(game.Players) >= 2 {
					gamesMutex.Unlock()
					err := ws.WriteJSON(map[string]string{"error": "game full"})
					if err != nil {
						return
					}
					log.Printf("Attempt to join full game with ID: %s", gameID)
					continue
				}
				playerColor := toggleColor(game.Players[0].Color)
				player := &Player{Conn: ws, Color: playerColor}
				game.Players = append(game.Players, player)
				gamesMutex.Unlock()
				err := ws.WriteJSON(map[string]string{"status": "joined", "gameID": gameID, "color": playerColor.String()})
				if err != nil {
					return
				}
				log.Printf("Player joined game with ID: %s", gameID)
				broadcastGameState(gameID)
			} else {
				gamesMutex.Unlock()
				err := ws.WriteJSON(map[string]string{"error": "game not found"})
				if err != nil {
					return
				}
				log.Printf("Attempt to join non-existent game with ID: %s", gameID)
			}
		case "move":
			gameID := msg["gameID"]
			moveStr := msg["move"]
			log.Printf("Move attempted in game %s: %s", gameID, moveStr)

			gamesMutex.Lock()
			game, exists := games[gameID]
			if exists {
				if game.Game.Position().Turn() != getPlayerColor(ws, game) {
					err := ws.WriteJSON(map[string]string{"error": "not your turn"})
					if err != nil {
						return
					}
					gamesMutex.Unlock()
					log.Printf("Invalid move attempt: not player's turn in game %s", gameID)
					continue
				}

				err := game.Game.MoveStr(moveStr)
				if err != nil {
					err := ws.WriteJSON(map[string]string{"error": err.Error()})
					if err != nil {
						return
					}
					gamesMutex.Unlock()
					log.Printf("Invalid move in game %s: %s", gameID, moveStr)
					continue
				}
				log.Printf("Move made in game %s: %s", gameID, moveStr)
				broadcastGameState(gameID)
				gamesMutex.Unlock()
			} else {
				gamesMutex.Unlock()
				err := ws.WriteJSON(map[string]string{"error": "game not found"})
				if err != nil {
					return
				}
				log.Printf("Attempt to move in non-existent game with ID: %s", gameID)
			}
		default:
			log.Printf("Unknown action: %s", action)
		}
	}
}

func broadcastGameState(gameID string) {
	game := games[gameID]
	game.Lock()
	defer game.Unlock()

	status := "ongoing"
	if game.Game.Outcome() != chess.NoOutcome {
		if game.Game.Method() == chess.Checkmate {
			status = "checkmate"
		} else if game.Game.Method() == chess.Stalemate {
			status = "stalemate"
		} else if game.Game.Method() == chess.InsufficientMaterial {
			status = "draw"
		}
	}

	state := map[string]string{
		"status": status,
		"fen":    game.Game.Position().String(),
	}

	for _, player := range game.Players {
		err := player.Conn.WriteJSON(state)
		if err != nil {
			log.Println("Write error:", err)
		}
	}

	log.Printf("Game state broadcast for game ID %s: %s", gameID, status)
}

func removePlayer(ws *websocket.Conn) {
	gamesMutex.Lock()
	defer gamesMutex.Unlock()

	for gameID, game := range games {
		game.Lock()
		for i, player := range game.Players {
			if player.Conn == ws {
				game.Players = append(game.Players[:i], game.Players[i+1:]...)
				log.Printf("Player removed from game ID %s", gameID)
				break
			}
		}
		if len(game.Players) == 0 {
			delete(games, gameID)
			log.Printf("Game ID %s deleted", gameID)
		}
		game.Unlock()
	}
}
