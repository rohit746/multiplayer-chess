package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebSocket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handleConnections))
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			return
		}
	}(ws)

	// Test game creation
	msg := map[string]string{"action": "create"}
	err = ws.WriteJSON(msg)
	if err != nil {
		return
	}

	_, response, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var res map[string]string
	err = json.Unmarshal(response, &res)
	if err != nil {
		return
	}
	gameID := res["gameID"]
	assert.NotEmpty(t, gameID, "Game ID should not be empty")

	// Test joining the game
	msg = map[string]string{"action": "join", "gameID": gameID}
	err = ws.WriteJSON(msg)
	if err != nil {
		return
	}

	_, response, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	err = json.Unmarshal(response, &res)
	if err != nil {
		return
	}
	assert.Equal(t, "joined", res["status"])

	// Test making a valid move
	msg = map[string]string{"action": "move", "gameID": gameID, "move": "e2e4"}
	err = ws.WriteJSON(msg)
	if err != nil {
		return
	}

	_, response, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	err = json.Unmarshal(response, &res)
	if err != nil {
		return
	}
	assert.Equal(t, "ongoing", res["status"])
	assert.Contains(t, res["fen"], "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR", "Board should reflect the move e2e4")

	// Test making an invalid move
	msg = map[string]string{"action": "move", "gameID": gameID, "move": "e2e5"}
	err = ws.WriteJSON(msg)
	if err != nil {
		return
	}

	_, response, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	err = json.Unmarshal(response, &res)
	if err != nil {
		return
	}
	assert.Equal(t, "invalid move", res["error"])
}
