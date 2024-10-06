import React, { useState } from "react";
import "./../../styles/AskAI.css";
import { GoogleGenerativeAI } from "@google/generative-ai";
import ChatbotAnimation from "./ChatbotAnimation";

// Initialize the GoogleGenerativeAI Client
const genAI = new GoogleGenerativeAI(process.env.REACT_APP_GEMINI_KEY);
const model = genAI.getGenerativeModel({ model: "gemini-pro" }); // Adjust based on your desired model

const ChatBot = () => {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [animationVisible, setAnimationVisible] = useState(true);

  const sendMessage = async () => {
    if (!input) return;

    setMessages([...messages, { sender: "user", text: input }]);
    setLoading(true);
    setAnimationVisible(false);

    try {
      // Modify the prompt to pre-prompt the model
      const prompt = `
            You are tasked to answer any user question that is appropriate and not intended in a harmful or inappropriate way. If you find any, please reject with a polite message.
            Answer in a 50-100 words and concise manner, unless specified otherwise.
            User question: ${input}`;

      const result = await model.generateContent(prompt);
      const response = await result.response;
      let chatbotResponse =
        (await response.text()) || "Sorry, I didn't understand that.";

      // Apply formatting if needed
      chatbotResponse = formatResponse(chatbotResponse);

      // Append the bot's response to the chat
      setMessages((prevMessages) => [
        ...prevMessages,
        { sender: "bot", text: chatbotResponse },
      ]);
    } catch (error) {
      console.error("Error communicating with chatbot:", error);
      setMessages((prevMessages) => [
        ...prevMessages,
        { sender: "bot", text: "Error: Couldn't fetch response." },
      ]);
    }

    setLoading(false);
    setInput(""); // Clear the input field
  };

  const formatResponse = (response) => {
    let formatted = response.trim();
    formatted = formatted.replace(/\*\*(.*?)\*\*/g, "<b>$1</b>");
    const lines = formatted.split("\n");
    let result = "";
    let listStack = [];

    lines.forEach((line) => {
      if (line.startsWith("* ")) {
        const text = line.substring(2).trim();
        const indentLevel = line.search(/\S/);
        const currentLevel = Math.floor(indentLevel / 4);

        while (listStack.length > currentLevel) {
          result += "</ul>";
          listStack.pop();
        }

        if (listStack.length < currentLevel) {
          result += "<ul>";
          listStack.push(currentLevel);
        }

        result += `<li>${text}</li>`;
      } else {
        while (listStack.length > 0) {
          result += "</ul>";
          listStack.pop();
        }
        result += `<p>${line}</p>`;
      }
    });

    while (listStack.length > 0) {
      result += "</ul>";
      listStack.pop();
    }

    return result;
  };

  return (
    <div className="askai">
      {animationVisible && (
        <div className="animation-placeholder">
          <ChatbotAnimation />
          <p className="welcome-message">
            Welcome! Ask me anything, and let's dive into learning together.
          </p>
        </div>
      )}
      <div className="chat-window">
        {messages.map((msg, idx) => (
          <div
            key={idx}
            className={msg.sender === "user" ? "user-message" : "bot-message"}
          >
            {msg.sender === "user" ? (
              msg.text
            ) : (
              <div dangerouslySetInnerHTML={{ __html: msg.text }} />
            )}
          </div>
        ))}
      </div>

      {loading && <p>Loading...</p>}

      <div style={{ display: "flex", alignItems: "center", marginTop: "1rem" }}>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={(e) => e.key === "Enter" && sendMessage()}
          placeholder="Explain the overall stucture of the Indian..."
        />
        <button className="button-ai" onClick={sendMessage} disabled={loading}>
          Send
        </button>
      </div>
    </div>
  );
};

export default ChatBot;
