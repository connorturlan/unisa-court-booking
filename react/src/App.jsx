import { useState } from "react";
import styles from "./App.module.scss";
import RadioCard from "./components/RadioCard/RadioCard";
import DateCard from "./components/DateCard/DateCard";

function App() {
  const [count, setCount] = useState(0);

  const onSubmit = (event) => {
    event.preventDefault();

    const formData = new FormData(event.target);

    console.log(formData.get("1"));
  };

  return (
    <>
      <div className={styles.App}>
        <h1>Court Booking</h1>
        <form onSubmit={onSubmit}>
          <RadioCard group={1} value={1}>
            <DateCard title="Date 1">details about date one</DateCard>
          </RadioCard>

          <RadioCard group={1} value={2}>
            <DateCard title="Date 2">details about date two</DateCard>
          </RadioCard>

          <input type="submit" value="book" />
        </form>
      </div>
    </>
  );
}

export default App;
