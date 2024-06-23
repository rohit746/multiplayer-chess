// App.tsx
import React, { useEffect, useState, useCallback } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import ChessBoard from "./components/ChessBoard";
import * as ChessJS from "chess.js"; // Import Chess from chess.js

interface ResponseMessage {
  color?: string;
  gameID?: string;
  status?: string;
  fen?: string;
  error?: string;
}

const App: React.FC = () => {
  const [socketUrl] = useState("ws://localhost:8080/ws");
  const { sendMessage, lastMessage, readyState } = useWebSocket(socketUrl);
  const [gameID, setGameID] = useState<string>("");
  const [inputGameID, setInputGameID] = useState<string>(""); // State to store user input for game ID
  const [joiningGameID, setJoiningGameID] = useState<string>(""); // State to store game ID to be sent to server
  const [color, setColor] = useState<"white" | "black" | "">("");
  const [game, setGame] = useState<ChessJS.Chess>(new ChessJS.Chess()); // Initialize game as Chess instance
  const [gameStatus, setGameStatus] = useState<string>("");

  const handleCreateGame = useCallback(() => {
    const message = JSON.stringify({ action: "create" });
    sendMessage(message);
  }, [sendMessage]);

  const handleJoinGame = useCallback(() => {
    if (inputGameID.trim() !== "") {
      setJoiningGameID(inputGameID.trim()); // Set joiningGameID state with user input
    }
  }, [inputGameID]);

  const handleMove = useCallback(
    (from: string, to: string) => {
      if (game) {
        const chessMove = game.move({ from, to, promotion: "q" }); // Always promote to queen

        if (chessMove) {
          const updatedGame = new ChessJS.Chess(game.fen());
          setGame(updatedGame);

          const moveNotation = chessMove.san;

          const message = JSON.stringify({
            action: "move",
            gameID,
            move: moveNotation,
          });
          sendMessage(message);

          return true;
        }
      }

      return false;
    },
    [game, gameID, sendMessage]
  );

  useEffect(() => {
    if (lastMessage !== null) {
      try {
        const parsedMessage: ResponseMessage = JSON.parse(lastMessage.data);
        if (parsedMessage.error) {
          console.error("Server error:", parsedMessage.error);
          return;
        }
        if (parsedMessage.gameID) {
          setGameID(parsedMessage.gameID);
        }
        if (parsedMessage.color) {
          setColor(parsedMessage.color === "w" ? "white" : "black");
        }
        if (parsedMessage.status) {
          setGameStatus(parsedMessage.status);
        }
        if (parsedMessage.fen) {
          const updatedGame = new ChessJS.Chess(parsedMessage.fen);
          setGame(updatedGame);
        }
      } catch (error) {
        console.error("Failed to parse message", error);
      }
    }
  }, [lastMessage]);

  useEffect(() => {
    if (joiningGameID !== "") {
      const message = JSON.stringify({ action: "join", gameID: joiningGameID });
      sendMessage(message);
    }
  }, [joiningGameID, sendMessage]);

  const connectionStatus = {
    [ReadyState.CONNECTING]: "Connecting",
    [ReadyState.OPEN]: "Open",
    [ReadyState.CLOSING]: "Closing",
    [ReadyState.CLOSED]: "Closed",
    [ReadyState.UNINSTANTIATED]: "Uninstantiated",
  }[readyState];

  return (
    <main className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-4">Chess Game</h1>
        <div className="flex space-x-4 mb-4">
          <button
            onClick={handleCreateGame}
            disabled={readyState !== ReadyState.OPEN}
            className="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50 hover:bg-blue-600"
          >
            Create Game
          </button>
          <input
            type="text"
            value={inputGameID}
            onChange={(e) => setInputGameID(e.target.value)}
            placeholder="Enter Game ID"
            className="px-4 py-2 border rounded w-40"
          />
          <button
            onClick={handleJoinGame}
            disabled={
              readyState !== ReadyState.OPEN || inputGameID.trim() === ""
            }
            className="px-4 py-2 bg-green-500 text-white rounded disabled:opacity-50 hover:bg-green-600"
          >
            Join Game
          </button>
        </div>
        <p className="mb-2">WebSocket status: {connectionStatus}</p>
        {gameID && (
          <div>
            <p>Game ID: {gameID}</p>
            <p>Your color: {color}</p>
            <p>Game status: {gameStatus}</p>
          </div>
        )}
      </div>
      {gameID && (
        <ChessBoard
          game={game}
          boardOrientation={color === "black" ? "black" : "white"}
          onMove={handleMove}
        />
      )}
    </main>
  );
};

export default App;
