import React, { useMemo, useState, useCallback } from "react";
import { Chessboard } from "react-chessboard";
import { Chess } from "chess.js";
import { Square } from "react-chessboard/dist/chessboard/types";

function ChessBoard() {
  const [game, setGame] = useState(new Chess());
  const [possibleMoves, setPossibleMoves] = useState<{
    [square: string]: "move" | "capture";
  }>({});
  const [selectedPiece, setSelectedPiece] = useState<Square | null>(null);

  // Create audio elements for each sound
  const moveAudio = useMemo(() => new Audio("/move.mp3"), []);
  const captureAudio = useMemo(() => new Audio("/capture.mp3"), []);
  const checkAudio = useMemo(() => new Audio("/check.mp3"), []);

  const playSound = useCallback(
    (soundType: "move" | "capture" | "check") => {
      switch (soundType) {
        case "move":
          moveAudio.play();
          break;
        case "capture":
          captureAudio.play();
          break;
        case "check":
          checkAudio.play();
          break;
      }
    },
    [moveAudio, captureAudio, checkAudio]
  );

  function makeAMove(from: Square, to: Square) {
    const gameCopy = new Chess(game.fen());
    const move = gameCopy.move({
      from,
      to,
      promotion: "q", // always promote to queen for simplicity
    });

    if (move === null) return false;

    setGame(gameCopy);
    setPossibleMoves({});
    setSelectedPiece(null);

    // Play sound based on move type
    if (move.captured) {
      playSound("capture");
    } else {
      playSound("move");
    }

    // Check for check
    if (gameCopy.inCheck()) {
      playSound("check");
    }

    return true;
  }

  function onSquareClick(square: Square) {
    const piece = game.get(square);

    // If there's a selected piece and it's the same as the clicked piece, deselect it
    if (selectedPiece === square) {
      setSelectedPiece(null);
      setPossibleMoves({});
      return;
    }

    // If no piece is selected or the clicked piece is of the current player's color, select it
    if (!selectedPiece || (piece && piece.color === game.turn())) {
      setSelectedPiece(square);
      const moves = game.moves({ square, verbose: true });
      const newPossibleMoves: { [square: string]: "move" | "capture" } = {};
      moves.forEach((move) => {
        newPossibleMoves[move.to] = move.captured ? "capture" : "move";
      });
      setPossibleMoves(newPossibleMoves);
    } else if (selectedPiece) {
      // If a piece is selected and it's not the same as the clicked piece, try to move it
      if (makeAMove(selectedPiece, square)) {
        return; // Move was successful
      }
    }
  }

  function onDrop(sourceSquare: Square, targetSquare: Square): boolean {
    return makeAMove(sourceSquare, targetSquare);
  }

  const customPieces = useMemo(() => {
    const pieces = [
      "wP",
      "wN",
      "wB",
      "wR",
      "wQ",
      "wK",
      "bP",
      "bN",
      "bB",
      "bR",
      "bQ",
      "bK",
    ];
    const pieceComponents: {
      [key: string]: React.FC<{ squareWidth: number }>;
    } = {};
    pieces.forEach((piece) => {
      pieceComponents[piece] = ({ squareWidth }) => (
        <div
          style={{
            width: squareWidth,
            height: squareWidth,
            backgroundImage: `url(/${piece}.png)`,
            backgroundSize: "100%",
          }}
        />
      );
    });
    return pieceComponents;
  }, []);

  const squareStyles = useMemo(() => {
    const styles: { [key: string]: React.CSSProperties } = {};

    // Add gradient for the king in check
    if (game.inCheck()) {
      const kingSquare = game.board().reduce((acc, row, i) => {
        const kingIndex = row.findIndex(
          (p) => p && p.type === "k" && p.color === game.turn()
        );
        return kingIndex !== -1
          ? `${String.fromCharCode(97 + kingIndex)}${8 - i}`
          : acc;
      }, "");
      if (kingSquare) {
        styles[kingSquare] = {
          background:
            "radial-gradient(circle, rgba(255,0,0,0.3) 0%, rgba(255,0,0,0.1) 50%, transparent 70%)",
        };
      }
    }

    // Highlight selected piece
    if (selectedPiece) {
      styles[selectedPiece] = {
        ...styles[selectedPiece],
        backgroundColor: "rgba(255, 255, 0, 0.4)", // Yellow highlight for selected piece
      };
    }

    // Add indicators for possible moves and captures
    Object.entries(possibleMoves).forEach(([square, moveType]) => {
      if (moveType === "move") {
        styles[square] = {
          ...styles[square],
          background: styles[square]?.background
            ? `${styles[square].background}, radial-gradient(circle at center, rgba(0,0,0,.2) 19%, transparent 20%)`
            : "radial-gradient(circle at center, rgba(0,0,0,.2) 19%, transparent 20%)",
          backgroundSize: styles[square]?.backgroundSize || "100%, 50% 50%",
          backgroundPosition:
            styles[square]?.backgroundPosition || "center, center",
          backgroundRepeat: "no-repeat, no-repeat",
        };
      } else if (moveType === "capture") {
        styles[square] = {
          ...styles[square],
          background: styles[square]?.background
            ? `${styles[square].background}, radial-gradient(circle at center, rgba(0,0,0,0) 0%, rgba(0,0,0,0) 39%, rgba(0,0,0,0.2) 40%, rgba(0,0,0,0.2) 59%, rgba(0,0,0,0) 60%)`
            : "radial-gradient(circle at center, rgba(0,0,0,0) 0%, rgba(0,0,0,0) 39%, rgba(0,0,0,0.2) 40%, rgba(0,0,0,0.2) 59%, rgba(0,0,0,0) 60%)",
          backgroundSize: styles[square]?.backgroundSize || "100%, 50% 50%",
          backgroundPosition:
            styles[square]?.backgroundPosition || "center, center",
          backgroundRepeat: "no-repeat, no-repeat",
        };
      }
    });

    return styles;
  }, [game, possibleMoves, selectedPiece]);

  return (
    <div className="w-full max-w-full sm:w-[90vw] md:w-[70vw] lg:w-[50vw] xl:w-[40vw] mx-auto my-12 p-4">
      <Chessboard
        position={game.fen()}
        onPieceDrop={onDrop}
        onSquareClick={onSquareClick}
        boardOrientation="white"
        customBoardStyle={{
          borderRadius: "4px",
          boxShadow: "0 2px 10px rgba(0, 0, 0, 0.5)",
        }}
        customDarkSquareStyle={{ backgroundColor: "#779952" }}
        customLightSquareStyle={{ backgroundColor: "#edeed1" }}
        customPieces={customPieces}
        customSquareStyles={squareStyles}
      />
    </div>
  );
}

export default ChessBoard;
