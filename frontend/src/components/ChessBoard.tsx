// ChessBoard.tsx
import React, { useMemo, useState, useCallback } from "react";
import { Chessboard } from "react-chessboard";
import { Chess, Square } from "chess.js";
import { BoardOrientation } from "react-chessboard/dist/chessboard/types";

interface ChessBoardProps {
  game: Chess;
  boardOrientation: BoardOrientation;
  onMove: (from: Square, to: Square) => boolean;
}

function ChessBoard({ game, boardOrientation, onMove }: ChessBoardProps) {
  const [selectedPiece, setSelectedPiece] = useState<Square | null>(null);

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

  const isPlayerTurn = useCallback(() => {
    return (
      (game.turn() === "w" && boardOrientation === "white") ||
      (game.turn() === "b" && boardOrientation === "black")
    );
  }, [game, boardOrientation]);

  function onSquareClick(square: Square) {
    if (!isPlayerTurn()) return;

    const piece = game.get(square);

    if (selectedPiece === square) {
      setSelectedPiece(null);
      return;
    }

    if (!selectedPiece || (piece && piece.color === game.turn())) {
      setSelectedPiece(square);
    } else if (selectedPiece) {
      const moveResult = onMove(selectedPiece, square);
      if (moveResult) {
        setSelectedPiece(null);
        const move = game.history({ verbose: true }).pop();
        if (move) {
          if (move.captured) {
            playSound("capture");
          } else {
            playSound("move");
          }
          if (game.inCheck()) {
            playSound("check");
          }
        }
      }
    }
  }

  function onDrop(sourceSquare: Square, targetSquare: Square): boolean {
    if (!isPlayerTurn()) return false;

    const piece = game.get(sourceSquare);
    if (!piece || piece.color !== game.turn()) return false;

    const moveResult = onMove(sourceSquare, targetSquare);
    if (moveResult) {
      const move = game.history({ verbose: true }).pop();
      if (move) {
        if (move.captured) {
          playSound("capture");
        } else {
          playSound("move");
        }
        if (game.inCheck()) {
          playSound("check");
        }
      }
    }
    return moveResult;
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
            "radial-gradient(circle, rgba(255,0,0,0.3) 0%, rgba(255,0,0,0.1) 50%, transparent 70%)", // Red gradient for check
        };
      }
    }

    if (game.isCheckmate()) {
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
          background: "rgba(255, 0, 0, 0.7)", // Dark red color for checkmate
        };
      }
    }

    if (selectedPiece && isPlayerTurn()) {
      styles[selectedPiece] = {
        ...styles[selectedPiece],
        backgroundColor: "rgba(255, 255, 0, 0.4)", // Yellow highlight for selected piece
      };

      const moves = game.moves({ square: selectedPiece, verbose: true });
      moves.forEach((move) => {
        if (move.captured) {
          styles[move.to] = {
            ...styles[move.to],
            background: styles[move.to]?.background
              ? `${
                  styles[move.to].background
                }, radial-gradient(circle at center, rgba(0,0,0,0) 0%, rgba(0,0,0,0) 39%, rgba(0,0,0,0.2) 40%, rgba(0,0,0,0.2) 59%, rgba(0,0,0,0) 60%)`
              : "radial-gradient(circle at center, rgba(0,0,0,0) 0%, rgba(0,0,0,0) 39%, rgba(0,0,0,0.2) 40%, rgba(0,0,0,0.2) 59%, rgba(0,0,0,0) 60%)",
            backgroundSize: styles[move.to]?.backgroundSize || "100%, 50% 50%",
            backgroundPosition:
              styles[move.to]?.backgroundPosition || "center, center",
            backgroundRepeat: "no-repeat, no-repeat",
          };
        } else {
          styles[move.to] = {
            ...styles[move.to],
            background: styles[move.to]?.background
              ? `${
                  styles[move.to].background
                }, radial-gradient(circle at center, rgba(0,0,0,.2) 19%, transparent 20%)`
              : "radial-gradient(circle at center, rgba(0,0,0,.2) 19%, transparent 20%)",
            backgroundSize: styles[move.to]?.backgroundSize || "100%, 50% 50%",
            backgroundPosition:
              styles[move.to]?.backgroundPosition || "center, center",
            backgroundRepeat: "no-repeat, no-repeat",
          };
        }
      });
    }

    return styles;
  }, [game, selectedPiece, isPlayerTurn]);

  return (
    <div className="w-full max-w-full sm:w-[90vw] md:w-[70vw] lg:w-[50vw] xl:w-[40vw] mx-auto my-12 p-4">
      <Chessboard
        position={game.fen()}
        onPieceDrop={onDrop}
        onSquareClick={onSquareClick}
        boardOrientation={boardOrientation}
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
