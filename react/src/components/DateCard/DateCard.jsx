import styles from "./DateCard.module.scss";

function DateCard({ title, children }) {
  return (
    <div className={styles.DateCard}>
      <h2 className={styles.DateCard_Title}>{title}</h2>
      <p className={styles.DateCard_Content}>{children}</p>
    </div>
  );
}

export default DateCard;
