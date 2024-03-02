import { useState } from "react";
import styles from "./App.module.scss";
import RadioCard from "./components/RadioCard/RadioCard";

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
            <p>Date 2</p>
          </RadioCard>

          <RadioCard group={1} value={2}>
            <p>Date 2</p>
          </RadioCard>

          <input type="submit" value="book" />
        </form>
      </div>
    </>
  );
}

export default App;
