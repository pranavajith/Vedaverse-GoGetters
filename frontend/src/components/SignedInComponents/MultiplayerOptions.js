import React, { useState } from "react";
import "./../../styles/MultiplayerOptions.css";

const MultiplayerOptions = ({ onCreateGame, onJoinGame }) => {
  const [status, setStatus] = useState("");

  const handleCreateGame = async () => {
    // Simulate API call to create game lobby
    setStatus("Creating game...");
    await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate API delay

    const dummyGameId = Math.floor(Math.random() * 1000); // Dummy game ID
    setStatus(`Game created. Waiting for opponent to join. Game ID: ${dummyGameId}`);
    onCreateGame(dummyGameId); // Simulate game creation
  };

  const handleJoinGame = async () => {
    // Simulate fetching available games
    setStatus("Searching for available games...");
    await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate API delay

    const availableGames = [101, 102, 103]; // Dummy available game IDs
    if (availableGames.length > 0) {
      const gameToJoin = availableGames[0]; // Automatically join the first game for simplicity
      setStatus(`Joined Game ID: ${gameToJoin}. Starting game...`);
      onJoinGame(gameToJoin); // Simulate game joining
    } else {
      setStatus("No available games to join.");
    }
  };

  return (
    <div className="cloud-container">
      <button className="button create-game" onClick={handleCreateGame}>
      <i className="fas fa-plus-circle"></i> Create a Game
      </button>
      <button className="button join-game" onClick={handleJoinGame}>
      <i className="fas fa-sign-in-alt"></i> Join a Game
      </button>
      <p className="status-message">{status}</p>
    </div>
  );
};

export default MultiplayerOptions;
