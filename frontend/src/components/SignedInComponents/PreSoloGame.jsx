import React, { useState, useEffect } from "react";
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
  const [headings, setHeadings] = useState([]);
  const [currentHeadingIndex, setCurrentHeadingIndex] = useState(0);
  const [quizData, setQuizData] = useState([]);
  const [roundLoading, setRoundLoading] = useState(false);

  const handleChange = (e) => {
    setTopic(e.target.value); // Update topic state on input change
  };

  // Fetch three short headings to cover the topic
  const generateHeadings = async (topic) => {
    const prompt = `
      Generate three short headings, each not more than 3 words, that can be used to comprehensively teach someone about the topic: "${topic}".
      Format the headings as follows:
      ["Heading 1", "Heading 2", "Heading 3"]
    `;
    try {
      const result = await model.generateContent(prompt);
      const response = result.response;
      let rawHeadings = await response.text();
      return JSON.parse(rawHeadings);
    } catch (error) {
      console.error("Error generating headings:", error);
      return [];
    }
  };

  // Fetch quiz questions for the current heading
  const generateQuizQuestions = async (heading) => {
    const prompt = `
      Generate a quiz with multiple levels for the topic: "${heading}".
      Format the questions as follows:
      [
        {
          "number": 1,
          "levelName": "Intro to ${heading}",
          "videoUrl": "",
          "questionType": "flashcard",
          "questionData": [
            {
              "question": "relevant question",
              "options": [
                { "value": "Option 1", "correctStatus": false },
                { "value": "Option 2", "correctStatus": false },
                { "value": "Option 3", "correctStatus": true },
                { "value": "Option 4", "correctStatus": false }
              ]
            }
          ]
        },
        {
          "number": 2,
          "levelName": "In Depth",
          "videoUrl": "",
          "questionType": "flashcard",
          "questionData": [
            {
              "question": "relevant question",
              "options": [
                { "value": "Option 1", "correctStatus": false },
                { "value": "Option 2", "correctStatus": false },
                { "value": "Option 3", "correctStatus": true },
                { "value": "Option 4", "correctStatus": false }
              ]
            }
          ]
        }
      ]
      Only generate the data. Do not give any other explanation or comments.
    `;
    try {
      const result = await model.generateContent(prompt);
      const response = result.response;
      let rawQuiz = response.text();
      return JSON.parse(rawQuiz);
    } catch (error) {
      console.error("Error generating quiz questions:", error);
      return [];
    }
  };

  // Generate the paragraph and keywords, then append the TypeGame question
  const generateTypeGame = async (heading) => {
    const paragraphPrompt = `
      Generate a 50-word paragraph on the topic: "${heading}". 
      Use very relevant and important words in this paragraph.
    `;

    try {
      // Fetch paragraph from Gemini
      const paragraphResult = await model.generateContent(paragraphPrompt);
      const paragraphResponse = paragraphResult.response;
      const paragraphText = await paragraphResponse.text();

      // Generate keywords based on the paragraph
      const keywordPrompt = `
        Use the following paragraph to generate 5 keywords and choices. 
        For each keyword, provide 3 options:
        Paragraph: "${paragraphText}"
        Format the output as:
        [
          {
            "word": "important word",
            "choices": ["wrong word 1", "important word", "wrong word 2"]
          },
          ...
        ]
      `;
      const keywordResult = await model.generateContent(keywordPrompt);
      const keywordResponse = keywordResult.response;
      const keywordsArray = await keywordResponse.text();

      // Create the TypeGame data structure
      const typeGameData = {
        number: 3,
        levelName: "TypeGame",
        videoUrl: "https://cdn-icons-mp4.flaticon.com/512/8617/8617218.mp4",
        questionType: "TypeGame",
        questionData: {
          text: paragraphText,
          keywords: JSON.parse(keywordsArray),
        },
      };

      // Append to quizData
      setQuizData((prevQuizData) => [...prevQuizData, typeGameData]);
    } catch (error) {
      console.error("Error generating TypeGame data:", error);
    }
  };

 // Submit form to start the game
const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true); // Start loading when the form is submitted
  
    // Fetch headings based on the entered topic
    const generatedHeadings = await generateHeadings(topic);
    setHeadings(generatedHeadings); // Save the generated headings
    setLoading(false);
    setSubmitted(true); // Move to the next phase
  
    // Start fetching the first round of questions
    setRoundLoading(true);
  
    // Fetch quiz questions
    const quizQuestions = await generateQuizQuestions(generatedHeadings[0]);
  
    // Generate TypeGame data for the first heading
    await generateTypeGame(generatedHeadings[0]);
  
    // Combine quiz questions and TypeGame data
    const combinedQuizData = [...quizQuestions, quizData[quizData.length - 1]];
  
    // Set combined quiz data
    setQuizData(combinedQuizData);
    setRoundLoading(false);
  };
  
  // Move to the next heading and fetch its quiz questions
  const nextRound = async () => {
    const nextHeadingIndex = currentHeadingIndex + 1;
    if (nextHeadingIndex < headings.length) {
      setRoundLoading(true);
      const quizQuestions = await generateQuizQuestions(headings[nextHeadingIndex]);
      setQuizData(quizQuestions);

      // Generate TypeGame for the next heading
      await generateTypeGame(headings[nextHeadingIndex]);
      setCurrentHeadingIndex(nextHeadingIndex);
      setRoundLoading(false);
    }
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
          {loading && <p>Generating headings, please wait...</p>}
        </div>
      ) : roundLoading ? (
        <div className="round-loading-container">
          <p>Round Loading for {headings[currentHeadingIndex]}, please wait...</p>
        </div>
      ) : (
        <SoloGame
          topic={headings[currentHeadingIndex]}
          quizData={quizData}
          onNextRound={nextRound}
        />
      )}
    </>
  );
};

export default PreSoloGame;
