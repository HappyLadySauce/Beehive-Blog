"use client";

import { useCallback, useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { History, Loader2, RotateCcw, Save, Trash2 } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import { createContentVersion, deleteContentVersion, listContentVersions, restoreContentVersion } from "@/lib/api/contents";
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
  const [deleting, setDeleting] = useState<number | null>(null);
  const [restoring, setRestoring] = useState<number | null>(null);
  const [snapshotName, setSnapshotName] = useState("");
  const [snapshotSummary, setSnapshotSummary] = useState("");
  const [expanded, setExpanded] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<VersionItem | null>(null);
  const [restoreTarget, setRestoreTarget] = useState<VersionItem | null>(null);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);

  const load = useCallback(async () => {
    await Promise.resolve();
    setLoading(true);
    try {
      const result = await listContentVersions(contentId);
      setVersions(result.items);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setLoading(false);
    }
  }, [contentId]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void load();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [load]);

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

  async function remove(versionNumber: number) {
    if (deleting !== null) return;
    setDeleting(versionNumber);
    setMessage(null);
    try {
      await deleteContentVersion(contentId, versionNumber);
      setVersions((current) => current.filter((version) => version.version_number !== versionNumber));
      setExpanded((current) => (current === versionNumber ? null : current));
      setMessage({ tone: "success", text: "版本已删除。" });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setDeleting(null);
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
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 8 }}>
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
                <div className={styles.versionActions}>
                  <button
                    className="secondary-button"
                    disabled={restoring !== null || deleting !== null}
                    type="button"
                    onClick={() => setRestoreTarget(version)}
                  >
                    {restoring === version.version_number ? <Loader2 aria-hidden className="spin" size={12} /> : <RotateCcw aria-hidden size={12} />}
                    回滚到此版本
                  </button>
                  <button
                    aria-label={`删除版本 ${version.version_number}`}
                    className="secondary-button icon-button"
                    disabled={restoring !== null || deleting !== null}
                    type="button"
                    onClick={() => setDeleteTarget(version)}
                  >
                    {deleting === version.version_number ? <Loader2 aria-hidden className="spin" size={12} /> : <Trash2 aria-hidden size={12} />}
                  </button>
                </div>
              </div>
              {expanded === version.version_number && (
                <div className={styles.versionPreview}>
                  <div className={styles.versionPreviewTitle}>{version.title}</div>
                  {version.excerpt && <div className={styles.versionPreviewExcerpt}>{version.excerpt}</div>}
                  <pre className={styles.versionPreviewBody}>
                    {version.body || "(无正文)"}
                  </pre>
                </div>
              )}
            </li>
          ))}
        </ul>
      )}
      {restoreTarget ? (
        <RestoreConfirmDialog
          restoring={restoring === restoreTarget.version_number}
          version={restoreTarget}
          onCancel={() => setRestoreTarget(null)}
          onConfirm={() => {
            const versionNumber = restoreTarget.version_number;
            setRestoreTarget(null);
            void restore(versionNumber);
          }}
        />
      ) : null}
      {deleteTarget ? (
        <DeleteVersionDialog
          deleting={deleting === deleteTarget.version_number}
          version={deleteTarget}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={() => {
            const versionNumber = deleteTarget.version_number;
            setDeleteTarget(null);
            void remove(versionNumber);
          }}
        />
      ) : null}
    </div>
  );
}

function versionDisplayName(version: VersionItem) {
  return version.snapshot_type === "auto" ? "自动保存" : version.name || `v${version.version_number}`;
}

function RestoreConfirmDialog(props: {
  restoring: boolean;
  version: VersionItem;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  const versionLabel = versionDisplayName(props.version);

  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-labelledby="restore-version-title" aria-modal="true" className={styles.modal} role="alertdialog">
        <h3 id="restore-version-title">确认回滚版本</h3>
        <p>
          将用「{versionLabel}」的标题、正文和摘要覆盖当前内容。Slug、状态、标签和分类不会改变，回滚前会自动保存当前内容为快照。
        </p>
        <div className={styles.modalActions}>
          <button className="secondary-button" disabled={props.restoring} type="button" onClick={props.onCancel}>
            取消
          </button>
          <button className="danger-button" disabled={props.restoring} type="button" onClick={props.onConfirm}>
            {props.restoring ? <Loader2 aria-hidden className="spin" size={14} /> : <RotateCcw aria-hidden size={14} />}
            确认回滚
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function DeleteVersionDialog(props: {
  deleting: boolean;
  version: VersionItem;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  const versionLabel = versionDisplayName(props.version);

  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-labelledby="delete-version-title" aria-modal="true" className={styles.modal} role="alertdialog">
        <h3 id="delete-version-title">确认删除版本</h3>
        <p>确认删除「{versionLabel}」？删除后该版本不能再用于回滚。</p>
        <div className={styles.modalActions}>
          <button className="secondary-button" disabled={props.deleting} type="button" onClick={props.onCancel}>
            取消
          </button>
          <button className="danger-button" disabled={props.deleting} type="button" onClick={props.onConfirm}>
            {props.deleting ? <Loader2 aria-hidden className="spin" size={14} /> : <Trash2 aria-hidden size={14} />}
            删除版本
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}
