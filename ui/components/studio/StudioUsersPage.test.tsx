import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { resetUsersPageModuleStateForTests, StudioUsersPage } from "./StudioUsersPage";

const listUsers = vi.hoisted(() => vi.fn());
const createUser = vi.hoisted(() => vi.fn());
const updateUser = vi.hoisted(() => vi.fn());
const deleteUser = vi.hoisted(() => vi.fn());
const uploadLocalAttachment = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/users", () => ({
  createUser,
  deleteUser,
  listUsers,
  updateUser
}));

vi.mock("@/lib/api/attachments", () => ({
  attachmentContentUrl: (id: number) => `/api/bff/attachments/${id}/content`,
  uploadLocalAttachment
}));

const users = {
  items: [
    {
      id: 1,
      username: "admin",
      email: "admin@example.com",
      nickname: "Admin",
      phone: null,
      role: "admin",
      status: "active",
      avatar_attachment_id: null,
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ],
  total: 1,
  page: 1,
  page_size: 20
};

function renderUsersPage() {
  return render(
    <ToastProvider>
      <StudioUsersPage />
    </ToastProvider>
  );
}

describe("StudioUsersPage", () => {
  beforeEach(() => {
    resetUsersPageModuleStateForTests();
    listUsers.mockReset();
    createUser.mockReset();
    updateUser.mockReset();
    deleteUser.mockReset();
    uploadLocalAttachment.mockReset();
    listUsers.mockResolvedValue(users);
  });

  it("loads the user list once on mount", async () => {
    renderUsersPage();
    await waitFor(() => expect(screen.getByText("admin")).toBeInTheDocument());
    expect(listUsers).toHaveBeenCalledTimes(1);
  });

  it("refetches only the user list when changing status filter", async () => {
    renderUsersPage();
    await waitFor(() => expect(screen.getByText("admin")).toBeInTheDocument());
    const callsAfterMount = listUsers.mock.calls.length;

    fireEvent.click(screen.getByRole("combobox", { name: "按状态筛选" }));
    fireEvent.click(screen.getByRole("option", { name: "活跃" }));

    await waitFor(() =>
      expect(listUsers).toHaveBeenLastCalledWith(
        expect.objectContaining({ page: 1, page_size: 20, status: "active" })
      )
    );
    expect(listUsers.mock.calls.length).toBeGreaterThan(callsAfterMount);
  });

  it("debounces search before calling the user list API", async () => {
    renderUsersPage();
    await waitFor(() => expect(screen.getByText("admin")).toBeInTheDocument());
    const callsAfterMount = listUsers.mock.calls.length;

    fireEvent.change(screen.getByLabelText("搜索用户名或邮箱"), { target: { value: "adm" } });

    await waitFor(
      () =>
        expect(listUsers).toHaveBeenLastCalledWith(
          expect.objectContaining({ search: "adm", page: 1, page_size: 20 })
        ),
      { timeout: 2000 }
    );
    expect(listUsers.mock.calls.length).toBe(callsAfterMount + 1);
  });
});
