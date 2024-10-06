import React, { useState } from "react";
import "./../../styles/UserHomePage.css";
import PreSoloGame from "./PreSoloGame";
import MultiplayerOptions from "./MultiplayerOptions";

export const UserHomePage = () => {
  const [display, setDisplay] = useState(0);
  const [gameId, setGameId] = useState(null);

  const onClick1 = () => setDisplay(1); // Single player
  const onClick2 = () => setDisplay(2); // Multiplayer

  const handleCreateGame = (createdGameId) => {
    // Set game ID when created
    setGameId(createdGameId);
    // Optionally route to another page when game is created
    // Route to waiting room, etc.
  };

  const handleJoinGame = (joinedGameId) => {
    // Set game ID after joining
    setGameId(joinedGameId);
    // Route to the game page when game starts
  };
  return (
    <>
      {display === 0 && (
        <div className="button-container-2">
          <button className="big-button single-player" onClick={onClick1}>
            Single Player
          </button>
          <button className="big-button multiplayer" onClick={onClick2}>
            Multiplayer
          </button>
        </div>
      )}
      {display === 1 && <PreSoloGame />}
      {display === 2 && (
        <MultiplayerOptions
          onCreateGame={handleCreateGame}
          onJoinGame={handleJoinGame}
        />
      )}
      {gameId && <p>Game is active with ID: {gameId}</p>}
    </>
  );
};

export default UserHomePage;
