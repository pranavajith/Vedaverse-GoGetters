import React, { useContext } from 'react';
import '../../styles/TaskScreen.css';
import QuestionSlider from '../general-components/QuestionSlider';
import TypeGame from '../SignedInComponents/TypeGame';

const TaskScreen = ({ level, handleReturn, onRoundComplete }) => {

  const handleComplete = async () => {
    onRoundComplete();
    handleReturn();
  };

  const renderGame = () => {
    console.log("Here is Level: ",level)
    console.log("Here is Type:", level.questionType)
    switch (level.questionType) {
      case 'flashcard':
        return (
          <QuestionSlider
            display_questions={level.questionData}
            onComplete={handleComplete}
            handleQuizReturn={handleReturn}
          />
        );
      case 'TypeGame':
        return (
          <TypeGame 
            displayData={level.questionData} 
            onComplete={handleComplete} 
            handleIncompleteReturn={handleReturn} 
          />
        );
      default:
        return <div>Unsupported game type</div>;
    }
  };

  return (
    <div className="task-screen-container">
      {renderGame()}
    </div>
  );
};

export default TaskScreen;
