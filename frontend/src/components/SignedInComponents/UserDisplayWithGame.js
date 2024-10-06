import UserGameplay from "./UserGameplay";
import "./../../styles/UserDisplayWithGame.css";
import "./../../styles/UserGameplay.css";
import React, { useState, useEffect } from "react";
import TaskScreen from "../GameComponents/TaskScreen";

const UserDisplayWithGame = ({ topic, quizData }) => {
  const [currentLevel, setCurrentLevel] = useState(null);
  const [quizDataWithStatus, setQuizDataWithStatus] = useState([]);

  // Initialize quiz data with status when the component mounts
  useEffect(() => {
    const initializedQuizData = quizData.map((data, index) => ({
      ...data,
      status: index === 0 ? "unlocked" : "locked", // First element unlocked, others locked
    }));
    setQuizDataWithStatus(initializedQuizData);
  }, [quizData]);

  const handleLevelClick = (level) => {
    setCurrentLevel(level);
  };

  const handleReturn = () => {
    setCurrentLevel(null);
  };

  // Function to handle completing a round
  const completeRound = (roundIndex) => {
    // Update the status of the next round to "unlocked"
    if (roundIndex + 1 < quizDataWithStatus.length) {
      setQuizDataWithStatus((prevData) =>
        prevData.map((data, index) =>
          index === roundIndex + 1 ? { ...data, status: "unlocked" } : data
        )
      );
    }
  };

  return (
    <div className="right-side-display">
      {currentLevel ? (
        <TaskScreen
          level={currentLevel}
          handleReturn={handleReturn}
          onRoundComplete={completeRound}
        />
      ) : (
        <UserGameplay
          inputLevels={quizDataWithStatus}
          topic={topic}
          handleLevelClick={handleLevelClick}
        />
      )}
    </div>
  );
};

export { UserDisplayWithGame };
