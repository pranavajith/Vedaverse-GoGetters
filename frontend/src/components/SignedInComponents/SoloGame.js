import React from "react";
import { UserProgress } from "./UserProgress";
import "./../../styles/SoloGame.css";
import { UserDisplayWithGame } from "./UserDisplayWithGame";

export const SoloGame = ({ topic, quizData }) => {
  console.log(topic);
  console.log(quizData);
  return (
    <div className="user-homepage-container">
      {/* UserProgress will remain on top */}
      <UserProgress />

      {/* A container to vertically align the two UserDisplayWithGame components */}
      <div className="game-display-container">
        <UserDisplayWithGame topic="Intro" quizData={quizData[0]} />
        <UserDisplayWithGame topic="Advanced" quizData={quizData[1]} />
      </div>
    </div>
  );
};

export default SoloGame;
