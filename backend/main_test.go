package main

import (
	"github.com/notnil/chess"
	"testing"
)

func TestMoveValidation(t *testing.T) {
	game := chess.NewGame()

	err := game.MoveStr("e2e4")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	err = game.MoveStr("e7e5")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	err = game.MoveStr("e2e5")
	if err == nil {
		t.Fatalf("expected error, but got none")
	}
}
