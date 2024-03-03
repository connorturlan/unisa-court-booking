import styles from "./App.module.scss";
import RadioCard from "./components/RadioCard/RadioCard";
import DateCard from "./components/DateCard/DateCard";
import UnisaTitle from "./components/UnisaTitle/UnisaTitle";
import { useEffect, useState } from "react";
import { getCookie, setCookie } from "./utils/cookies";

function App() {
  const [sessionsJson, setSessionsJson] = useState({});
  const [sessionsHtml, setSessionsHtml] = useState({});

  const [isLoading, setLoadingState] = useState(true);
  const [hidden, setHidden] = useState(getCookie("isHidden") == "true");

  const sendSessionRequest = async () => {
    const res = await fetch(
      "https://4wcagmhkc0.execute-api.ap-southeast-2.amazonaws.com/Prod/sessions",
      { headers: { Accept: "application/json" } }
    );

    const jsonBody = await res.json();

    setSessionsJson(jsonBody);
  };

  const sendBookingReservation = async (event, formData) => {
    const data = Array.from(formData.keys()).map((key) => {
      const value = formData.get(key);
      return `${key}-${value}`;
    });
    data.pop();

    const email = formData.get("email");
    console.log("data:", email, data);

    if (email == "") {
      window.alert("Missing email.");
      return;
    }

    const body = {
      uid: email,
      sessions: data,
    };

    const bodyJson = JSON.stringify(body);

    const res = await fetch(
      "https://4wcagmhkc0.execute-api.ap-southeast-2.amazonaws.com/Prod/booking",
      {
        method: "POST",
        headers: { Accept: "application/json" },
        body: bodyJson,
      }
    );

    if (res.status != 202) {
      window.alert(res.status);
    }
    event.target.reset();
    window.alert("Your booking has been submitted!");

    setHidden(true);
    setCookie("isHidden", "true", 7);
  };

  const onSubmit = (event) => {
    event.preventDefault();
    const formData = new FormData(event.target);
    sendBookingReservation(event, formData);
  };

  const parseSessionData = (json) => {
    const sessions = json["Sessions"];

    if (!sessions) {
      return <p>error!</p>;
    }

    const elements = sessions.map((sessionGroup, group) => {
      if (sessionGroup.Length <= 0) {
        return <p>error!</p>;
      }

      return (
        <section key={group} className={styles.Form_Date}>
          {sessionGroup.map((dateObject, dateId) => {
            const date = dateObject["Date"];
            const details = dateObject["Details"] || date;
            const stock = dateObject["Available"] || "-1";

            if (!date || !details || !stock) {
              return <p>error</p>;
            }

            const dateTime = new Date(date);
            const dateString = `${dateTime.toLocaleDateString("en-au", {
              weekday: "short",
            })} ${dateTime.toLocaleDateString()}`;

            return (
              <RadioCard key={dateId} group={group} value={dateId}>
                <DateCard title={dateString}>
                  <p>{details}</p>
                  <p>Available: {stock}</p>
                </DateCard>
              </RadioCard>
            );
          })}
        </section>
      );
    });

    return elements;
  };

  useEffect(() => {
    sendSessionRequest();
  }, []);

  useEffect(() => {
    setSessionsHtml(parseSessionData(sessionsJson));
    setLoadingState(false);
  }, [sessionsJson]);

  return (
    <>
      <div className={styles.App}>
        <div className={styles.App_Container}>
          <div className={styles.Container}>
            <div className={styles.Info}>
              <UnisaTitle />
              <h2>Player Booking Register</h2>
              <p>
                Due to the high demand we encounter during our come and try
                sessions, we are trialing a booking system for our players.
              </p>
              <p>
                We are implementing this system in order to provide our players
                the safest possible forum to be able to enjoy the game we all
                love. We ask kindly for your patience during this busy time.
              </p>
              <p>
                <i>
                  Please select one of each of the following dates, followed by
                  your email. You are welcome to resubmit your reservation at
                  any time, with the risk that there may not be availablity.
                </i>
              </p>
              <h3 className={styles.Info_Notice} hidden>
                Booking is currently closed until Monday 4th
              </h3>
            </div>
            <form onSubmit={onSubmit} className={styles.Form}>
              <div className={styles.Form_Sections}>
                {isLoading ? (
                  <h2 className={styles.Form_Sections}>Loading...</h2>
                ) : (
                  <>{sessionsHtml}</>
                )}
              </div>
              <section className={styles.Form_Email}>
                <label htmlFor="email">Email: </label>
                <input type="email" name="email" id="email" />
              </section>
              <input
                className={styles.Form_Submit}
                type="submit"
                value="Reserve"
              />
            </form>
            <img className={styles.Info_Unisa} src="unisaSport.png" alt="" />
          </div>
        </div>
      </div>
    </>
  );
}

export default App;
