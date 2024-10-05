import React from "react";
import { UserProgress } from "./UserProgress";
import "./../../styles/SoloGame.css";
import { UserDisplayWithGame } from "./UserDisplayWithGame";

export const SoloGame = ({ topic, quizData }) => {
  console.log(topic);
  console.log(quizData);
  return (
    <div className="user-homepage-container">
      {/* <h2 className="game-topic">Learning Topic: {topic}</h2> */}
      <UserProgress />
      <UserDisplayWithGame topic={topic[0]} quizData={quizData[0]} />
      <UserDisplayWithGame topic={topic[1]} quizData={quizData[1]} />
      <UserDisplayWithGame topic={topic[2]} quizData={quizData[2]} />
    </div>
  );
};

export default SoloGame;
