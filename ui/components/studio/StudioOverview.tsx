import { Clock3, FileText, ShieldCheck } from "lucide-react";

import styles from "./Studio.module.css";
import { releaseChecks, studioMetrics } from "./studio-data";
import { StudioMetricCard } from "./StudioMetricCard";
import { StudioPanel } from "./StudioPanel";
import { StudioTopbar } from "./StudioTopbar";

export function StudioOverview() {
  return (
    <>
      <StudioTopbar
        actions={
          <button className="primary-button" disabled type="button">
            <FileText aria-hidden size={18} />
            新建内容
          </button>
        }
        description="管理公开内容、草稿队列、发布状态与后续 AI 审阅入口。"
        eyebrow="Owner workspace"
        title="内容管理与发布闸门"
      />

      <section className={styles.metricGrid} aria-label="Studio 指标">
        {studioMetrics.map((metric) => (
          <StudioMetricCard
            detail={metric.detail}
            icon={metric.icon}
            key={metric.label}
            label={metric.label}
            value={metric.value}
          />
        ))}
      </section>

      <div className={styles.workGrid}>
        <StudioPanel title="待处理">
          <div className={styles.panelHeader}>
            <div className="status-line">
              <Clock3 aria-hidden size={20} />
              <strong>Content API 接入前保持只读骨架</strong>
            </div>
          </div>
          <p>
            当前前端先完成身份闭环、Studio 信息架构和发布闸门骨架。内容写入、版本、关系和 AI 审阅会在
            content API 稳定后接入这里。
          </p>
        </StudioPanel>

        <StudioPanel title="发布检查">
          <ul className={styles.checkList}>
            {releaseChecks.map((check) => (
              <li className={styles.checkItem} key={check.label}>
                <span>{check.label}</span>
                <span className={`${styles.statusPill} ${check.state === "ready" ? styles.statusReady : styles.statusPending}`}>
                  {check.state === "ready" ? "Ready" : "Pending"}
                </span>
              </li>
            ))}
          </ul>
        </StudioPanel>
      </div>

      <StudioPanel title="权限状态" action={<ShieldCheck aria-hidden size={20} />}>
        <p>Studio 路由由 BFF Cookie 会话与服务端中间件优先拦截。客户端只展示当前状态，不持有 refresh token。</p>
      </StudioPanel>
    </>
  );
}
