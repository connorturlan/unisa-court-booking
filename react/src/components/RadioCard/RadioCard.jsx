import { useId } from "react";
import styles from "./RadioCard.module.scss";

function RadioCard({ group, value, children }) {
  const id = useId();

  return (
    <div className={styles.RadioCard}>
      <input
        className={styles.RadioCard_Input}
        type="radio"
        name={group}
        id={id}
        value={value}
        hidden
      />

      <label className={styles.RadioCard_Content} htmlFor={id}>
        {children}
      </label>
    </div>
  );
}

export default RadioCard;
