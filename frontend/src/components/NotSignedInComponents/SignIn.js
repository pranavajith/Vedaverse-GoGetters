import React, { useContext, useState } from "react";
import "./../../styles/SignIn.css";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { UserContext } from "../../context/UserContext";
import Banner from "../general-components/Banner";
import "./../../utils/i18n";
import { useTranslation } from "react-i18next";
import axios from "axios"; // Import axios
import { urlList } from "../../urls";

const SignIn = () => {
  const navigate = useNavigate();
  const { login } = useContext(UserContext);
  const location = useLocation();
  const bannerMessage = location.state?.message;

  const [formData, setFormData] = useState({
    username: "",
    password: "",
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevFormData) => ({
      ...prevFormData,
      [name]: value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault(); // Prevent form from submitting the default way
    login();
  };

  const { t } = useTranslation();

  return (
    <>
      {bannerMessage && <Banner message={bannerMessage} />}
      <div className="signin-contain">
        <div className="signin-container">
          <h1 className="signin-title">{t("wlcm")}</h1>
          <form className="signin-form" onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="username">
                <span role="img" aria-label="username">
                  👤
                </span>{" "}
                {t("username")}
              </label>
              <input
                type="text"
                id="username"
                name="username"
                value={formData.username}
                onChange={handleChange}
                placeholder={t("username")}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="password">
                <span role="img" aria-label="lock">
                  🔒
                </span>{" "}
                {t("pass")}
              </label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                placeholder={t("passh")}
                required
              />
            </div>
            <button type="submit" className="signin-button">
              <span role="img" aria-label="key">
                🔑
              </span>{" "}
              {t("login")}
            </button>
          </form>
          <div className="signin-footer">
            <p>
              {t("accna")}
              <Link to="/signup">{t("sup")}</Link>
            </p>
          </div>
        </div>
      </div>
    </>
  );
};

export default SignIn;
