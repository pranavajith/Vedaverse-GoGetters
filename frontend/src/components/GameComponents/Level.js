import React, { useContext } from "react";
import "../../styles/Level.css";
import { urlList } from "../../urls";
import { UserContext } from "../../context/UserContext";

const Level = ({ level, onClick }) => {
  console.log("here is the level:", level);
  const check = true;
  return (
    <div
      className={`level ${check ? "unlocked" : "locked"}`}
      onClick={check ? onClick : null}
    >
      <video autoPlay loop muted className="level-video">
        <source
          src="https://cdn-icons-mp4.flaticon.com/512/6844/6844338.mp4"
          type="video/mp4"
        />
      </video>

      {!check && (
        <div className="overlay-lock">
          <img src={urlList.WhiteLockUrl} alt="Locked" className="lock-icon" />
        </div>
      )}
    </div>
  );
};

export default Level;
