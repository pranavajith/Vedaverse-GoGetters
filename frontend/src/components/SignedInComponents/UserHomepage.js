import React, { useState } from "react";
import "./../../styles/UserHomePage.css";
import PreSoloGame from "./PreSoloGame";

export const UserHomePage = () => {
  const [display, setDisplay] = useState(0);
  const onClick1 = () => setDisplay(1);
  const onClick2 = () => setDisplay(2);
  return (
    <>
      {display === 0 && (
        <div className="button-container">
          <button className="big-button single-player" onClick={onClick1}>
            Single Player
          </button>
          <button className="big-button multiplayer" onClick={onClick2}>
            Multiplayer
          </button>
        </div>
      )}
      {display === 1 && <PreSoloGame />}
      {display === 2 && null}
    </>
  );
};

export default UserHomePage;
