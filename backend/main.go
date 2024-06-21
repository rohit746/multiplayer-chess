package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
	"github.com/segmentio/ksuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Game struct {
	sync.Mutex
	Game    *chess.Game
	Players []*websocket.Conn
}

var games = make(map[string]*Game)
var gamesMutex sync.Mutex

func main() {
	http.HandleFunc("/ws", handleConnections)
	log.Println("Server started on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	log.Println("New connection established from", r.RemoteAddr)

	defer func() {
		log.Println("Connection closed from", r.RemoteAddr)
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
			gamesMutex.Lock()
			games[gameID] = &Game{
				Game:    chess.NewGame(),
				Players: []*websocket.Conn{ws},
			}
			gamesMutex.Unlock()
			err := ws.WriteJSON(map[string]string{"status": "created", "gameID": gameID})
			if err != nil {
				return
			}
			log.Printf("Game created with ID: %s", gameID)
		case "join":
			gameID := msg["gameID"]
			gamesMutex.Lock()
			game, exists := games[gameID]
			if exists {
				game.Players = append(game.Players, ws)
				gamesMutex.Unlock()
				err := ws.WriteJSON(map[string]string{"status": "joined", "gameID": gameID})
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
		err := player.WriteJSON(state)
		if err != nil {
			log.Println("Write error:", err)
		}
	}

	log.Printf("Game state broadcast for game ID %s: %s", gameID, status)
}
