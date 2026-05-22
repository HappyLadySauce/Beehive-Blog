import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AnzhiyuCommentBox } from "./AnzhiyuCommentBox";

const createPublicComment = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/content", () => ({
  createPublicComment
}));

describe("AnzhiyuCommentBox", () => {
  beforeEach(() => {
    createPublicComment.mockReset();
  });

  it("submits an anonymous comment for moderation", async () => {
    createPublicComment.mockResolvedValue({ id: 9, status: "review" });

    render(<AnzhiyuCommentBox comments={[]} contentID={7} />);

    fireEvent.change(screen.getByPlaceholderText("欢迎留下宝贵的建议哟 ~"), { target: { value: "写得很好" } });
    fireEvent.change(screen.getByPlaceholderText("昵称 必填"), { target: { value: "读者" } });
    fireEvent.change(screen.getByPlaceholderText("邮箱 必填"), { target: { value: "reader@example.com" } });
    fireEvent.submit(screen.getByRole("button", { name: "发送" }).closest("form")!);

    await waitFor(() => expect(createPublicComment).toHaveBeenCalledWith(7, expect.objectContaining({ body: "写得很好" })));
    expect(await screen.findByText("评论已提交，审核通过后会显示。")).toBeInTheDocument();
  });
});
