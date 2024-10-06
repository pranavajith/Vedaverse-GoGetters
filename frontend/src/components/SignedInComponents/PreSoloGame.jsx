import React, { useState } from "react";
import SoloGame from "./SoloGame";
import { GoogleGenerativeAI } from "@google/generative-ai"; // Ensure you have this imported correctly
import "./../../styles/PreSoloGame.css";

// Initialize the GoogleGenerativeAI Client
const genAI = new GoogleGenerativeAI(process.env.REACT_APP_GEMINI_KEY);
const model = genAI.getGenerativeModel({ model: "gemini-pro" }); // Adjust based on your desired model

const PreSoloGame = () => {
  const [topic, setTopic] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [loading, setLoading] = useState(false);
  const [quizData, setQuizData] = useState([]);
  const [roundLoading, setRoundLoading] = useState(false);

  const handleChange = (e) => {
    setTopic(e.target.value); // Update topic state on input change
  };

  // Fetch quiz questions for the topic, including TypeGame data
  const generateQuizQuestions = async (topic) => {
    const prompt = `
        Generate 2 short headings (less than or equal to 3 words each) regarding ${topic}. Make sure these 2 points are such that it would help someone learn about ${topic} in a structured manner.
        DO NOT DISPLAY THE SHORT HEADINGS. ONLY KEEP THEM IN MIND.

      Generate a quiz with multiple levels for the topic: "${topic}".
      Format the questions as follows:
      [
      
      [
        {
          "number": 1,
          "levelName": // Intro to the first short heading,
          "questionType": "flashcard",
          "questionData": [
            {
              "question": "relevant question to the first short heading",
              "options": [
                { "value": "Option 1", "correctStatus": false },
                { "value": "Option 2", "correctStatus": false },
                { "value": "Option 3", "correctStatus": true },
                { "value": "Option 4", "correctStatus": false }
              ]
            }
          ]
            // Generate 2 such flashcards for the "Intro" level
        },
        {
          "number": 2,
          "levelName": "In Depth of the first short heading",
          "questionType": "flashcard",
          "questionData": [
            {
              "question": "relevant question to the first short heading",
              "options": [
                { "value": "Option 1", "correctStatus": false },
                { "value": "Option 2", "correctStatus": false },
                { "value": "Option 3", "correctStatus": true },
                { "value": "Option 4", "correctStatus": false }
              ]
            },
            // Generate 2 such flashcards for the "In Depth" level
          ]
        },
        {
          "number": 3,
          "levelName": "Final Round",
          "questionType": "TypeGame",
          "questionData": {
            "text": (Generate a 50-word paragraph on the first generated short heading. Use relevant and important words in the paragraph.),
            "keywords": [
              {
                "word": Choose an important word from the aobve paragraph,
                "choices": [Give 3 words, 2 of which are wrong, and the third is the chosen word above],
                // Make sure the choices are single words.
                // The chosen word (Z) MUST be present in the paragraph.
              },
              ...
            ]
          }
        }
      ],

      [
        {
          "number": 1,
          "levelName": // Intro to the second short heading,
          "videoUrl": "",
          "questionType": "flashcard",
          "questionData": [
            {
              "question": "relevant question to the second short heading",
              "options": [
                { "value": "Option 1", "correctStatus": false },
                { "value": "Option 2", "correctStatus": false },
                { "value": "Option 3", "correctStatus": true },
                { "value": "Option 4", "correctStatus": false }
              ]
            }
          ]
            // Generate 2 such flashcards for the "Intro" level
        },
        {
          "number": 2,
          "levelName": "In Depth of the second short heading",
          "videoUrl": "",
          "questionType": "flashcard",
          "questionData": [
            {
              "question": "relevant question to the second short heading",
              "options": [
                { "value": "Option 1", "correctStatus": false },
                { "value": "Option 2", "correctStatus": false },
                { "value": "Option 3", "correctStatus": true },
                { "value": "Option 4", "correctStatus": false }
              ]
            }
          ]
            // Generate 2 such flashcards for the "In Depth" level
        },
        {
          "number": 3,
          "levelName": "Final Round",
          "questionType": "TypeGame",
          "questionData": {
            "text": (Generate a 50-word paragraph on the second generated short heading. Use relevant and important words in the paragraph.),
            "keywords": [
              {
                "word": Choose an important word from the aobve paragraph,
                "choices": [Give 3 words, 2 of which are wrong, and the third is the chosen word above],
                // Make sure the choices are single words.
                // The chosen word (Z) MUST be present in the paragraph.
              },
              ...
            ]
          }
        }
      ],
        ]
      Only generate the data. Do not give any other explanation or comments.
    `;

    try {
      const result = await model.generateContent(prompt);
      const response = result.response;
      let rawQuiz = await response.text();
      return JSON.parse(rawQuiz);
    } catch (error) {
      console.error("Error generating quiz questions:", error);
      return [];
    }
  };

  // Submit form to start the game
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true); // Start loading when the form is submitted
  
    // Fetch quiz questions (including TypeGame data)
    const quizQuestions = await generateQuizQuestions(topic);

    // Set quiz data and transition to the game
    setQuizData(quizQuestions);
    setLoading(false);
    setSubmitted(true); // Move to the next phase
    setRoundLoading(false);
  };

  return (
    <>
      {!submitted ? (
        <div className="pre-solo-game-container">
          <form onSubmit={handleSubmit}>
            <label htmlFor="topic-input">What do you want to learn about?</label>
            <input
              id="topic-input"
              type="text"
              value={topic}
              onChange={handleChange}
              placeholder="Enter a topic"
              required
            />
            <button type="submit" disabled={loading}>
              {loading ? "Loading..." : "Start Game"}
            </button>
          </form>
          {loading && <p className="generating-message">Generating quiz, please wait...</p>}
        </div>
      ) : roundLoading ? (
        <div className="round-loading-container">
          <p>Generating questions, please wait...</p>
        </div>
      ) : (
        <SoloGame
          topic={topic}
          quizData={quizData}
        />
      )}
    </>
  );
};

export default PreSoloGame;
