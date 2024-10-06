import React, { useState } from "react";
import TopicInput from "./TopicInput"; // Import TopicInput component
import SoloGame from "./SoloGame"; // Import SoloGame component

const TopicSelection = () => {
  const [topic, setTopic] = useState("");
  const [isTopicSubmitted, setIsTopicSubmitted] = useState(false);

  const handleTopicSubmit = (submittedTopic) => {
    setTopic(submittedTopic);
    setIsTopicSubmitted(true);
  };

  return (
    <>
      {isTopicSubmitted ? (
        <SoloGame topic={topic} /> // Render SoloGame with the selected topic
      ) : (
        <TopicInput onSubmit={handleTopicSubmit} /> // Render TopicInput
      )}
    </>
  );
};

export default TopicSelection;
