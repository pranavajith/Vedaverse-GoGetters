import React from "react";
import "./../../styles/UserGameplay.css";
import LevelMap from "../GameComponents/LevelMap";

const UserGameplay = ({ inputLevels, topic, handleLevelClick }) => {
  return (
    <div>
      <LevelMap
        levels={inputLevels}
        onLevelClick={handleLevelClick}
        levelText={topic}
      />
    </div>
  );
};

export default UserGameplay;
