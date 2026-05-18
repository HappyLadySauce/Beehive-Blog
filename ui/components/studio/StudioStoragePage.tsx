"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import {
  Activity,
  CheckCircle2,
  Database,
  HardDrive,
  Loader2,
  Pencil,
  Plus,
  Save,
  Star,
  Trash2,
  X
} from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import {
  checkStorageMount,
  createStorageMount,
  deleteStorageMount,
  disableStorageMount,
  enableStorageMount,
  listFileDrivers,
  listStorageMounts,
  updateStorageMount
} from "@/lib/api/storage";
import type { DriverResponse, JsonObject, StorageMountResponse } from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioSelect } from "./StudioSelect";
import { StudioTopbar } from "./StudioTopbar";

type StorageData = {
  drivers: DriverResponse[];
  mounts: StorageMountResponse[];
};

const emptyData: StorageData = {
  drivers: [],
  mounts: []
};

let storageLoadRequest: Promise<StorageData> | null = null;

function loadStorageData() {
  if (!storageLoadRequest) {
    storageLoadRequest = Promise.all([listFileDrivers(), listStorageMounts()])
      .then(([drivers, mounts]) => ({
        drivers: drivers.items,
        mounts: mounts.items
      }))
      .finally(() => {
        storageLoadRequest = null;
      });
  }
  return storageLoadRequest;
}

export function StudioStoragePage() {
  const [activeSection, setActiveSection] = useState<"mounts" | "drivers">("mounts");
  const [data, setData] = useState<StorageData>(emptyData);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingMount, setEditingMount] = useState<StorageMountResponse | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<StorageMountResponse | null>(null);

  const [formDriver, setFormDriver] = useState("local");
  const [formName, setFormName] = useState("");
  const [formPath, setFormPath] = useState("");
  const [formRemark, setFormRemark] = useState("");
  const [formOrder, setFormOrder] = useState("0");
  const [formDefault, setFormDefault] = useState(false);
  const [formConfig, setFormConfig] = useState(formatJson(defaultConfigForDriver("local")));

  const driverOptions = useMemo(
    () => data.drivers.map((driver) => ({ value: driver.name, label: `${driver.display_name} (${driver.name})` })),
    [data.drivers]
  );
  const driverMap = useMemo(() => new Map(data.drivers.map((driver) => [driver.name, driver])), [data.drivers]);

  const refreshStorage = useCallback(async () => {
    try {
      const [drivers, mounts] = await Promise.all([listFileDrivers(), listStorageMounts()]);
      setData({ drivers: drivers.items, mounts: mounts.items });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    }
  }, []);

  useEffect(() => {
    let active = true;
    loadStorageData()
      .then((result) => {
        if (active) setData(result);
      })
      .catch((error: unknown) => {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, []);

  function openCreate() {
    const nextDriver = data.drivers[0]?.name ?? "local";
    setEditingMount(null);
    setFormDriver(nextDriver);
    setFormName("");
    setFormPath(defaultMountPath(nextDriver));
    setFormRemark("");
    setFormOrder("0");
    setFormDefault(data.mounts.length === 0);
    setFormConfig(formatJson(defaultConfigForDriver(nextDriver)));
    setMessage(null);
    setShowForm(true);
  }

  function openEdit(mount: StorageMountResponse) {
    setEditingMount(mount);
    setFormDriver(mount.driver_name);
    setFormName(mount.name);
    setFormPath(mount.mount_path);
    setFormRemark(mount.remark ?? "");
    setFormOrder(String(mount.order_index));
    setFormDefault(mount.is_default);
    setFormConfig(formatJson(mount.config));
    setMessage(null);
    setShowForm(true);
  }

  function closeForm() {
    setShowForm(false);
    setEditingMount(null);
  }

  function onDriverChange(nextDriver: string) {
    setFormDriver(nextDriver);
    if (!editingMount) {
      setFormPath(defaultMountPath(nextDriver));
      setFormConfig(formatJson(defaultConfigForDriver(nextDriver)));
    }
  }

  async function onSubmitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage(null);

    if (formName.trim() === "") {
      setMessage({ tone: "error", text: "存储名称不能为空。" });
      return;
    }
    if (!editingMount && formPath.trim() === "") {
      setMessage({ tone: "error", text: "挂载路径不能为空。" });
      return;
    }

    const parsedConfig = parseConfig(formConfig);
    if (!parsedConfig.ok) {
      setMessage({ tone: "error", text: parsedConfig.error });
      return;
    }

    const orderIndex = Number.parseInt(formOrder, 10);
    setSaving(true);
    try {
      if (editingMount) {
        await updateStorageMount(editingMount.id, {
          config: parsedConfig.value,
          is_default: formDefault,
          name: formName.trim(),
          order_index: Number.isFinite(orderIndex) ? orderIndex : 0,
          remark: formRemark.trim() || null
        });
        setMessage({ tone: "success", text: "存储实例已更新。" });
      } else {
        await createStorageMount({
          config: parsedConfig.value,
          driver_name: formDriver,
          is_default: formDefault,
          mount_path: formPath.trim(),
          name: formName.trim(),
          order_index: Number.isFinite(orderIndex) ? orderIndex : 0,
          remark: formRemark.trim() || null
        });
        setMessage({ tone: "success", text: "存储实例已创建。" });
      }
      closeForm();
      await refreshStorage();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function toggleMount(mount: StorageMountResponse) {
    setSaving(true);
    setMessage(null);
    try {
      if (mount.disabled) {
        await enableStorageMount(mount.id);
        setMessage({ tone: "success", text: `${mount.name} 已启用。` });
      } else {
        await disableStorageMount(mount.id);
        setMessage({ tone: "success", text: `${mount.name} 已禁用。` });
      }
      await refreshStorage();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function markDefault(mount: StorageMountResponse) {
    setSaving(true);
    setMessage(null);
    try {
      await updateStorageMount(mount.id, { is_default: true });
      setMessage({ tone: "success", text: `${mount.name} 已设为默认存储。` });
      await refreshStorage();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function runHealthCheck(mount: StorageMountResponse) {
    setSaving(true);
    setMessage(null);
    try {
      const result = await checkStorageMount(mount.id);
      setMessage({
        tone: result.status === "work" ? "success" : "error",
        text: result.error ? `${mount.name} 健康检查失败：${result.error}` : `${mount.name} 健康检查通过。`
      });
      await refreshStorage();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    setSaving(true);
    setMessage(null);
    try {
      await deleteStorageMount(deleteTarget.id);
      setMessage({ tone: "success", text: `${deleteTarget.name} 已删除。` });
      setDeleteTarget(null);
      await refreshStorage();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <StudioTopbar
        actions={activeSection === "mounts" ? (
          <button className="primary-button" type="button" onClick={openCreate}>
            <Plus aria-hidden size={18} />
            添加存储
          </button>
        ) : null}
        description="管理存储实例与驱动模板；附件上传未指定存储时使用默认启用实例。"
        eyebrow="Storage management"
        title="存储管理"
      />

      <div className={styles.segmentedTabs} role="tablist" aria-label="存储页面分段">
        <button
          className={activeSection === "mounts" ? styles.segmentedTabActive : styles.segmentedTab}
          type="button"
          onClick={() => setActiveSection("mounts")}
        >
          存储实例
        </button>
        <button
          className={activeSection === "drivers" ? styles.segmentedTabActive : styles.segmentedTab}
          type="button"
          onClick={() => setActiveSection("drivers")}
        >
          文件驱动
        </button>
      </div>

      {message ? (
        <p className={`${styles.message} ${message.tone === "success" ? styles.messageSuccess : styles.messageError}`} role="alert">
          {message.text}
        </p>
      ) : null}

      {activeSection === "mounts" ? (
        <StudioPanel action={<Database aria-hidden size={22} />} title="存储实例">
          {loading ? (
            <div className={styles.emptyState} role="status">
              <Loader2 aria-hidden className="spin" size={24} />
              正在加载存储实例...
            </div>
          ) : data.mounts.length === 0 ? (
            <div className={styles.emptyState}>暂无存储实例。</div>
          ) : (
            <div className={styles.storageMountList}>
              {data.mounts.map((mount) => (
                <article className={styles.storageMountCard} key={mount.id}>
                  <div className={styles.storageMountHeader}>
                    <div>
                      <div className={styles.storageMountTitle}>
                        <strong>{mount.name}</strong>
                        <span className={styles.codePill}>{mount.mount_path}</span>
                      </div>
                      <div className={styles.metaRow}>
                        <StatusPill mount={mount} />
                        <span className={styles.codePill}>{mount.driver_name}</span>
                        {mount.is_default ? <span className={styles.statusPill}>默认</span> : null}
                      </div>
                    </div>
                    <div className={styles.tableActions}>
                      <button
                        className="secondary-button"
                        disabled={saving}
                        type="button"
                        onClick={() => openEdit(mount)}
                      >
                        <Pencil aria-hidden size={16} />
                        编辑
                      </button>
                      <button
                        className="secondary-button"
                        disabled={saving}
                        type="button"
                        onClick={() => runHealthCheck(mount)}
                      >
                        <Activity aria-hidden size={16} />
                        检查
                      </button>
                    </div>
                  </div>

                  <div className={styles.storageDetails}>
                    <span>排序 {mount.order_index}</span>
                    <span>更新 {formatDate(mount.updated_at)}</span>
                    {mount.last_checked_at ? <span>检查 {formatDate(mount.last_checked_at)}</span> : null}
                  </div>
                  {mount.remark ? <p className={styles.storageRemark}>{mount.remark}</p> : null}
                  {mount.last_error ? <p className={styles.storageError}>{mount.last_error}</p> : null}
                  <pre className={styles.codeBlock}>{formatJson(mount.config)}</pre>

                  <div className={styles.storageActions}>
                    <button className="secondary-button" disabled={saving} type="button" onClick={() => toggleMount(mount)}>
                      {mount.disabled ? <CheckCircle2 aria-hidden size={16} /> : <X aria-hidden size={16} />}
                      {mount.disabled ? "启用" : "禁用"}
                    </button>
                    <button
                      className="secondary-button"
                      disabled={saving || mount.is_default || mount.disabled}
                      type="button"
                      onClick={() => markDefault(mount)}
                    >
                      <Star aria-hidden size={16} />
                      设为默认
                    </button>
                    <button className="danger-button" disabled={saving} type="button" onClick={() => setDeleteTarget(mount)}>
                      <Trash2 aria-hidden size={16} />
                      删除
                    </button>
                  </div>
                </article>
              ))}
            </div>
          )}
        </StudioPanel>
      ) : null}

      {activeSection === "drivers" ? (
        <StudioPanel action={<HardDrive aria-hidden size={22} />} title="文件驱动">
          {loading ? (
            <div className={styles.emptyState} role="status">
              <Loader2 aria-hidden className="spin" size={24} />
              正在加载驱动...
            </div>
          ) : data.drivers.length === 0 ? (
            <div className={styles.emptyState}>暂无文件驱动。</div>
          ) : (
            <div className={styles.driverList}>
              {data.drivers.map((driver) => (
                <article className={styles.driverCard} key={driver.name}>
                  <div className={styles.storageMountTitle}>
                    <strong>{driver.display_name}</strong>
                    <span className={styles.codePill}>{driver.name}</span>
                  </div>
                  {driver.description ? <p>{driver.description}</p> : null}
                  <div className={styles.metaRow}>
                    <span className={driver.status === "active" ? styles.statusReady : styles.statusPending}>
                      {driver.status}
                    </span>
                    <span>{driverMountCount(driver.name, data.mounts)} 个实例</span>
                  </div>
                  <pre className={styles.codeBlock}>{formatJson(driver.capabilities)}</pre>
                </article>
              ))}
            </div>
          )}
        </StudioPanel>
      ) : null}

      {showForm &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modalWide} role="dialog">
              <div className={styles.modalHeader}>
                <div>
                  <h3>{editingMount ? "编辑存储实例" : "添加存储实例"}</h3>
                  <p>本地存储也是普通文件驱动实例，配置保存在数据库挂载项中。</p>
                </div>
                <button aria-label="关闭" className="icon-button" type="button" onClick={closeForm}>
                  <X aria-hidden size={18} />
                </button>
              </div>

              <form className={styles.formGrid} id="storage-mount-form" onSubmit={onSubmitForm}>
                <label className={styles.field}>
                  <span>驱动</span>
                  <StudioSelect
                    ariaLabel="驱动"
                    disabled={Boolean(editingMount)}
                    options={driverOptions}
                    value={formDriver}
                    onChange={onDriverChange}
                  />
                </label>

                <label className={styles.field}>
                  <span>名称</span>
                  <input aria-label="名称" value={formName} onChange={(event) => setFormName(event.target.value)} />
                </label>

                <label className={styles.field}>
                  <span>挂载路径</span>
                  <input
                    aria-label="挂载路径"
                    disabled={Boolean(editingMount)}
                    placeholder="/local"
                    value={formPath}
                    onChange={(event) => setFormPath(event.target.value)}
                  />
                </label>

                <label className={styles.field}>
                  <span>排序</span>
                  <input aria-label="排序" type="number" value={formOrder} onChange={(event) => setFormOrder(event.target.value)} />
                </label>

                <label className={styles.checkboxField}>
                  <input checked={formDefault} type="checkbox" onChange={(event) => setFormDefault(event.target.checked)} />
                  <span>设为默认存储</span>
                </label>

                <label className={styles.fieldFull}>
                  <span>备注</span>
                  <input aria-label="备注" value={formRemark} onChange={(event) => setFormRemark(event.target.value)} />
                </label>

                <label className={styles.fieldFull}>
                  <span>驱动配置 JSON</span>
                  <textarea
                    aria-label="驱动配置 JSON"
                    className={styles.textarea}
                    value={formConfig}
                    onChange={(event) => setFormConfig(event.target.value)}
                  />
                </label>

                {driverMap.get(formDriver)?.config_schema ? (
                  <div className={styles.fieldFull}>
                    <span className={styles.subsectionTitle}>配置 schema</span>
                    <pre className={styles.codeBlock}>{formatJson(driverMap.get(formDriver)?.config_schema ?? {})}</pre>
                  </div>
                ) : null}
              </form>

              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={closeForm}>
                  取消
                </button>
                <button className="primary-button" disabled={saving} form="storage-mount-form" type="submit">
                  {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
                  保存
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}

      {deleteTarget &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modal} role="dialog">
              <h3>删除存储实例</h3>
              <p>确认删除 {deleteTarget.name}？如果已有附件引用该存储，后端会拒绝删除。</p>
              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={() => setDeleteTarget(null)}>
                  取消
                </button>
                <button className="danger-button" disabled={saving} type="button" onClick={confirmDelete}>
                  删除
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}
    </>
  );
}

function StatusPill({ mount }: { mount: StorageMountResponse }) {
  if (mount.disabled) {
    return <span className={styles.statusPending}>已禁用</span>;
  }
  if (mount.status === "work") {
    return <span className={styles.statusReady}>正常</span>;
  }
  if (mount.status === "error") {
    return <span className={styles.messageError}>异常</span>;
  }
  return <span className={styles.statusPending}>未知</span>;
}

function defaultConfigForDriver(driverName: string): JsonObject {
  if (driverName === "local") return { root: "data/attachments" };
  if (driverName === "s3" || driverName === "oss") {
    return { bucket: "", upload_base_url: "", download_base_url: "" };
  }
  return {};
}

function defaultMountPath(driverName: string) {
  return `/${driverName || "storage"}`;
}

function parseConfig(input: string): { ok: true; value: JsonObject } | { ok: false; error: string } {
  try {
    const parsed = JSON.parse(input) as unknown;
    if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") {
      return { ok: false, error: "驱动配置必须是 JSON 对象。" };
    }
    return { ok: true, value: parsed as JsonObject };
  } catch {
    return { ok: false, error: "驱动配置不是合法 JSON。" };
  }
}

function formatJson(value: unknown) {
  return JSON.stringify(value ?? {}, null, 2);
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    month: "2-digit"
  }).format(new Date(value));
}

function driverMountCount(driverName: string, mounts: StorageMountResponse[]) {
  return mounts.filter((mount) => mount.driver_name === driverName).length;
}
