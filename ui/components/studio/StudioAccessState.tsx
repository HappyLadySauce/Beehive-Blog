import Link from "next/link";
import { ShieldAlert } from "lucide-react";

import styles from "./Studio.module.css";

type StudioAccessStateProps = {
  title: string;
  message: string;
  actionHref?: string;
  actionLabel?: string;
};

export function StudioAccessState({ actionHref, actionLabel, message, title }: StudioAccessStateProps) {
  return (
    <section className={styles.accessState} role="status">
      <ShieldAlert aria-hidden size={24} />
      <h1>{title}</h1>
      <p className={styles.muted}>{message}</p>
      {actionHref && actionLabel ? (
        <Link className="primary-button" href={actionHref}>
          {actionLabel}
        </Link>
      ) : null}
    </section>
  );
}
