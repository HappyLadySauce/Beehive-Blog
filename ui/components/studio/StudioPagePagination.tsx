"use client";

import styles from "./Studio.module.css";

export type PageToken = number | "ellipsis";

// buildPageNumbers returns a compact page sequence with ellipses.
// buildPageNumbers 返回带省略号的紧凑页码序列。
export function buildPageNumbers(page: number, totalPages: number): PageToken[] {
  if (totalPages <= 1) return [1];
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  const tokens: PageToken[] = [1];
  const left = Math.max(2, page - 1);
  const right = Math.min(totalPages - 1, page + 1);
  if (left > 2) tokens.push("ellipsis");
  for (let current = left; current <= right; current += 1) {
    tokens.push(current);
  }
  if (right < totalPages - 1) tokens.push("ellipsis");
  tokens.push(totalPages);
  return tokens;
}

export function StudioPagePagination(props: {
  disabled?: boolean;
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}) {
  const tokens = buildPageNumbers(props.page, props.totalPages);

  return (
    <nav aria-label="分页" className={`${styles.pagination} ${styles.paginationEnd}`}>
      <button
        className="secondary-button"
        disabled={props.disabled || props.page <= 1}
        type="button"
        onClick={() => props.onPageChange(props.page - 1)}
      >
        上一页
      </button>
      <div className={styles.pageNumberGroup} role="list">
        {tokens.map((token, index) =>
          token === "ellipsis" ? (
            <span key={`ellipsis-${index}`} className={styles.pageEllipsis} aria-hidden>
              …
            </span>
          ) : (
            <button
              key={token}
              aria-current={token === props.page ? "page" : undefined}
              aria-label={`第 ${token} 页`}
              className={token === props.page ? styles.pageNumberActive : styles.pageNumber}
              disabled={props.disabled}
              type="button"
              onClick={() => props.onPageChange(token)}
            >
              {token}
            </button>
          )
        )}
      </div>
      <span className={styles.pageTotalHint}>共 {props.totalPages} 页</span>
      <button
        className="secondary-button"
        disabled={props.disabled || props.page >= props.totalPages}
        type="button"
        onClick={() => props.onPageChange(props.page + 1)}
      >
        下一页
      </button>
    </nav>
  );
}
