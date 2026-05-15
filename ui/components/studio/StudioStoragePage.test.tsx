import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudioStoragePage } from "./StudioStoragePage";

const listFileDrivers = vi.hoisted(() => vi.fn());
const listStorageMounts = vi.hoisted(() => vi.fn());
const createStorageMount = vi.hoisted(() => vi.fn());
const updateStorageMount = vi.hoisted(() => vi.fn());
const enableStorageMount = vi.hoisted(() => vi.fn());
const disableStorageMount = vi.hoisted(() => vi.fn());
const checkStorageMount = vi.hoisted(() => vi.fn());
const deleteStorageMount = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/storage", () => ({
  checkStorageMount,
  createStorageMount,
  deleteStorageMount,
  disableStorageMount,
  enableStorageMount,
  listFileDrivers,
  listStorageMounts,
  updateStorageMount
}));

const drivers = {
  items: [
    {
      id: 1,
      name: "local",
      display_name: "Local Storage",
      description: "Server-local filesystem storage.",
      config_schema: { required: ["root"] },
      capabilities: { upload: true, download: true, delete: true },
      status: "active",
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

const mounts = {
  items: [
    {
      id: 10,
      driver_name: "local",
      mount_path: "/local",
      name: "Local Storage",
      remark: null,
      config: { root: "data/attachments" },
      order_index: 0,
      is_default: true,
      disabled: false,
      status: "work",
      last_checked_at: null,
      last_error: null,
      created_by: null,
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

describe("StudioStoragePage", () => {
  beforeEach(() => {
    listFileDrivers.mockReset();
    listStorageMounts.mockReset();
    createStorageMount.mockReset();
    updateStorageMount.mockReset();
    enableStorageMount.mockReset();
    disableStorageMount.mockReset();
    checkStorageMount.mockReset();
    deleteStorageMount.mockReset();

    listFileDrivers.mockResolvedValue(drivers);
    listStorageMounts.mockResolvedValue(mounts);
    createStorageMount.mockResolvedValue(mounts.items[0]);
    updateStorageMount.mockResolvedValue(mounts.items[0]);
    enableStorageMount.mockResolvedValue(mounts.items[0]);
    disableStorageMount.mockResolvedValue({ ...mounts.items[0], disabled: true });
    checkStorageMount.mockResolvedValue({ status: "work", checked: "2026-05-15T00:00:00Z" });
    deleteStorageMount.mockResolvedValue({});
  });

  it("loads drivers and storage mounts", async () => {
    render(<StudioStoragePage />);

    expect(screen.getByText("正在加载存储实例...")).toBeInTheDocument();
    await waitFor(() => expect(screen.getAllByText("Local Storage")).toHaveLength(2));
    expect(screen.getByText("/local")).toBeInTheDocument();
    expect(screen.getByText("默认")).toBeInTheDocument();
  });

  it("creates a local storage mount with JSON config", async () => {
    listStorageMounts.mockResolvedValueOnce({ items: [] });
    render(<StudioStoragePage />);
    await waitFor(() => expect(screen.getByRole("button", { name: "添加存储" })).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "添加存储" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "Uploads" } });
    fireEvent.change(screen.getByLabelText("挂载路径"), { target: { value: "/uploads" } });
    fireEvent.change(screen.getByLabelText("驱动配置 JSON"), { target: { value: '{ "root": "data/uploads" }' } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createStorageMount).toHaveBeenCalled());
    expect(createStorageMount.mock.calls[0][0]).toMatchObject({
      config: { root: "data/uploads" },
      driver_name: "local",
      mount_path: "/uploads",
      name: "Uploads"
    });
  });

  it("runs health checks and disables mounts", async () => {
    render(<StudioStoragePage />);
    await waitFor(() => expect(screen.getByText("/local")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "检查" }));
    await waitFor(() => expect(checkStorageMount).toHaveBeenCalledWith(10));

    fireEvent.click(screen.getByRole("button", { name: "禁用" }));
    await waitFor(() => expect(disableStorageMount).toHaveBeenCalledWith(10));
  });
});
