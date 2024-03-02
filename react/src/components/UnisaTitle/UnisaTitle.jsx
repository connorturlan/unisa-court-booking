import styles from "./UnisaTitle.module.scss";

function UnisaTitle() {
  return (
    <div className={styles.UnisaTitle}>
      <div className={styles.UnisaTitle_Image}>
        <img className={styles.UnisaTitle_Goats} src="goatsIcon.jpg" alt="" />
      </div>
      <h1 className={styles.UnisaTitle_Title}>UniSA Volleyball Club</h1>
    </div>
  );
}

export default UnisaTitle;
