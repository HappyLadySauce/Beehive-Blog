"use client";

import { useEffect, useState } from "react";
import { History, Loader2, RotateCcw, Save } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import { createContentVersion, listContentVersions, restoreContentVersion } from "@/lib/api/contents";
import type { ContentDetailResponse, VersionItem } from "@/lib/api/types";
import styles from "./Studio.module.css";

type Props = {
  contentId: number;
  onRestored: (detail: ContentDetailResponse) => void;
};

export function StudioContentVersionsPanel({ contentId, onRestored }: Props) {
  const [versions, setVersions] = useState<VersionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [restoring, setRestoring] = useState<number | null>(null);
  const [snapshotName, setSnapshotName] = useState("");
  const [snapshotSummary, setSnapshotSummary] = useState("");
  const [expanded, setExpanded] = useState<number | null>(null);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);

  async function load() {
    setLoading(true);
    try {
      const result = await listContentVersions(contentId);
      setVersions(result.items);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [contentId]);

  async function createSnapshot() {
    const name = snapshotName.trim();
    if (!name) {
      setMessage({ tone: "error", text: "版本名称不能为空。" });
      return;
    }
    setSaving(true);
    setMessage(null);
    try {
      const summary = snapshotSummary.trim() || undefined;
      await createContentVersion(contentId, { snapshot_type: "manual", name, change_summary: summary });
      setSnapshotName("");
      setSnapshotSummary("");
      setMessage({ tone: "success", text: "版本已保存。" });
      await load();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function restore(versionNumber: number) {
    if (restoring !== null) return;
    setRestoring(versionNumber);
    setMessage(null);
    try {
      const detail = await restoreContentVersion(contentId, versionNumber);
      setMessage({ tone: "success", text: `已回滚到版本 v${versionNumber}。` });
      onRestored(detail);
      await load();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setRestoring(null);
    }
  }

  const toggleExpand = (versionNumber: number) => {
    setExpanded((current) => (current === versionNumber ? null : versionNumber));
  };

  return (
    <div style={{ marginTop: 8 }}>
      {message && <div style={{ fontSize: 13, color: message.tone === "error" ? "#ef4444" : "#16a34a", marginBottom: 8 }}>{message.text}</div>}

      <div className={styles.versionSnapshotForm}>
        <input
          aria-label="版本名称"
          className={styles.versionSnapshotInput}
          maxLength={128}
          placeholder="版本名称"
          value={snapshotName}
          onChange={(event) => setSnapshotName(event.target.value)}
        />
        <input
          aria-label="版本说明"
          className={styles.versionSnapshotInput}
          maxLength={512}
          placeholder="版本说明（可选）"
          value={snapshotSummary}
          onChange={(event) => setSnapshotSummary(event.target.value)}
        />
        <button className={`secondary-button ${styles.versionSnapshotButton}`} disabled={saving} type="button" onClick={() => void createSnapshot()}>
          {saving ? <Loader2 aria-hidden className="spin" size={14} /> : <Save aria-hidden size={14} />}
          保存版本
        </button>
      </div>

      {loading ? (
        <div style={{ fontSize: 13, color: "#6b7280", padding: "8px 0" }}>
          <Loader2 aria-hidden className="spin" size={14} /> 正在加载版本历史...
        </div>
      ) : versions.length === 0 ? (
        <div style={{ fontSize: 13, color: "#6b7280", padding: "8px 0" }}>
          <History aria-hidden size={14} style={{ marginRight: 4 }} />
          暂无版本。内容变动保存后会生成自动保存版本，也可以手动保存命名版本。
        </div>
      ) : (
        <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
          {versions.map((version) => (
            <li key={version.id} style={{ borderBottom: "1px solid #e5e7eb", padding: "8px 0" }}>
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
                <button
                  style={{ background: "none", border: "none", cursor: "pointer", fontSize: 13, fontWeight: 500, textAlign: "left", padding: 0 }}
                  type="button"
                  onClick={() => toggleExpand(version.version_number)}
                >
                  {version.snapshot_type === "auto" ? "自动保存" : version.name || `v${version.version_number}`}
                  <span style={{ color: "#9ca3af", marginLeft: 8 }}>
                    {version.snapshot_type === "auto" ? "自动" : `v${version.version_number}`}
                  </span>
                  {version.change_summary ? ` — ${version.change_summary}` : ""}
                  <span style={{ color: "#9ca3af", marginLeft: 8 }}>
                    {new Date(version.created_at).toLocaleString()}
                  </span>
                </button>
                <button
                  className="secondary-button"
                  disabled={restoring !== null}
                  style={{ fontSize: 12 }}
                  type="button"
                  onClick={() => {
                    if (window.confirm("将用此版本快照的标题、正文和摘要覆盖当前内容。slug、状态、标签和分类不会改变。回滚前会自动保存当前内容为快照。确认回滚？")) {
                      void restore(version.version_number);
                    }
                  }}
                >
                  {restoring === version.version_number ? <Loader2 aria-hidden className="spin" size={12} /> : <RotateCcw aria-hidden size={12} />}
                  回滚到此版本
                </button>
              </div>
              {expanded === version.version_number && (
                <div style={{ marginTop: 8, fontSize: 13, background: "#f9fafb", borderRadius: 4, padding: "8px 12px" }}>
                  <div style={{ fontWeight: 600, marginBottom: 4 }}>{version.title}</div>
                  {version.excerpt && <div style={{ color: "#6b7280", marginBottom: 4, fontStyle: "italic" }}>{version.excerpt}</div>}
                  <pre style={{ whiteSpace: "pre-wrap", wordBreak: "break-word", fontSize: 12, margin: 0, maxHeight: 200, overflow: "auto" }}>
                    {version.body || "(无正文)"}
                  </pre>
                </div>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
