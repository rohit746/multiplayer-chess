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
		return true // Allow all connections
	},
}

var (
	games      = make(map[string]*Game)
	gamesMutex sync.Mutex
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	// Handle WebSocket communication
	for {
		var msg map[string]string
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		// Process WebSocket messages (e.g., game actions, moves)
		handleMessage(ws, msg)
	}
}

func handleMessage(ws *websocket.Conn, msg map[string]string) {
	// Implement your WebSocket message handling logic here
	log.Printf("Received message: %v", msg)

	// Example: Handle different message types (create, join, move)
	action := msg["action"]
	switch action {
	case "create":
		createGame(ws)
	case "join":
		joinGame(ws, msg["gameID"])
	case "move":
		makeMove(ws, msg["gameID"], msg["move"])
	default:
		log.Printf("Unknown action: %s", action)
	}
}

func createGame(ws *websocket.Conn) {
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

	// Notify the player about the game creation
	err := ws.WriteJSON(map[string]string{"status": "created", "gameID": gameID, "color": playerColor.String()})
	if err != nil {
		log.Println("Error sending game creation response:", err)
		return
	}

	log.Printf("Game created with ID: %s", gameID)
}

func joinGame(ws *websocket.Conn, gameID string) {
	gamesMutex.Lock()
	game, exists := games[gameID]
	if !exists {
		gamesMutex.Unlock()
		err := ws.WriteJSON(map[string]string{"error": "game not found"})
		if err != nil {
			log.Println("Error sending game not found response:", err)
		}
		log.Printf("Attempt to join non-existent game with ID: %s", gameID)
		return
	}

	if len(game.Players) >= 2 {
		gamesMutex.Unlock()
		err := ws.WriteJSON(map[string]string{"error": "game full"})
		if err != nil {
			log.Println("Error sending game full response:", err)
		}
		log.Printf("Attempt to join full game with ID: %s", gameID)
		return
	}

	playerColor := toggleColor(game.Players[0].Color)
	player := &Player{Conn: ws, Color: playerColor}
	game.Players = append(game.Players, player)
	gamesMutex.Unlock()

	// Notify the player about successfully joining the game
	err := ws.WriteJSON(map[string]string{"status": "joined", "gameID": gameID, "color": playerColor.String()})
	if err != nil {
		log.Println("Error sending game join response:", err)
		return
	}

	log.Printf("Player joined game with ID: %s", gameID)

	// Broadcast updated game state to all players
	broadcastGameState(gameID)
}

func makeMove(ws *websocket.Conn, gameID, moveStr string) {
	gamesMutex.Lock()
	game, exists := games[gameID]
	if !exists {
		gamesMutex.Unlock()
		err := ws.WriteJSON(map[string]string{"error": "game not found"})
		if err != nil {
			log.Println("Error sending game not found response:", err)
		}
		log.Printf("Attempt to move in non-existent game with ID: %s", gameID)
		return
	}

	if game.Game.Position().Turn() != getPlayerColor(ws, game) {
		gamesMutex.Unlock()
		err := ws.WriteJSON(map[string]string{"error": "not your turn"})
		if err != nil {
			log.Println("Error sending not your turn response:", err)
		}
		log.Printf("Invalid move attempt: not player's turn in game %s", gameID)
		return
	}

	err := game.Game.MoveStr(moveStr)
	if err != nil {
		gamesMutex.Unlock()
		err := ws.WriteJSON(map[string]string{"error": err.Error()})
		if err != nil {
			log.Println("Error sending move error response:", err)
		}
		log.Printf("Invalid move in game %s: %s", gameID, moveStr)
		return
	}

	gamesMutex.Unlock()

	log.Printf("Move made in game %s: %s", gameID, moveStr)

	// Broadcast updated game state to all players
	broadcastGameState(gameID)
}

func broadcastGameState(gameID string) {
	gamesMutex.Lock()
	game, exists := games[gameID]
	if !exists {
		gamesMutex.Unlock()
		log.Printf("Game not found when broadcasting game state for ID: %s", gameID)
		return
	}
	game.Lock()

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
			log.Println("Error broadcasting game state:", err)
		}
	}

	game.Unlock()
	gamesMutex.Unlock()

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
