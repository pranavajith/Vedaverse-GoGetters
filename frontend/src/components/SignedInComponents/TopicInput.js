// TopicSelect.js
import React, { useState } from "react";
import "./../../styles/TopicSelect.css"; // Import styles for TopicSelect

const TopicSelect = ({ onTopicSelect }) => {
  const [selectedTopic, setSelectedTopic] = useState("");

  const handleChange = (e) => {
    setSelectedTopic(e.target.value);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (selectedTopic) {
      onTopicSelect(selectedTopic); // Pass the selected topic back to the parent component
    }
  };

  return (
    <div className="topic-select-container">
      <h2 className="topic-select-title">Select a Topic to Learn About:</h2>
      <form onSubmit={handleSubmit} className="topic-select-form">
        <select
          value={selectedTopic}
          onChange={handleChange}
          className="topic-select-dropdown"
        >
          <option value="" disabled>
            Select a topic
          </option>
          <option value="Math">Math</option>
          <option value="Science">Science</option>
          <option value="History">History</option>
          {/* Add more topics as needed */}
        </select>
        <button type="submit" className="topic-select-button">
          Start Learning
        </button>
      </form>
    </div>
  );
};

export default TopicSelect;
