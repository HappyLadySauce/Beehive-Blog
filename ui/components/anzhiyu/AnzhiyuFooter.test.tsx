import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AnzhiyuFooter } from "./AnzhiyuFooter";

const stats = {
  articles: 12,
  notes: 4,
  projects: 3,
  views: 1877,
  tags: 16
};

describe("AnzhiyuFooter", () => {
  it("renders site, author, beian and statistics", () => {
    render(
      <AnzhiyuFooter
        author={{
          name: "Beehive",
          description: "个人博客与知识中台",
          bio: "持续整理公开文章、笔记和项目。",
          location: "广东 深圳",
          website: "https://example.com/me"
        }}
        site={{
          name: "Beehive Blog",
          url: "https://example.com",
          subtitle: "Beehive",
          description: "面向读者的公开知识库。",
          keywords: "blog, knowledge",
          logo_url: "",
          favicon_url: "",
          icp_beian: "湘ICP备2023015794号-2",
          police_beian: "公网安备 44030000000000号",
          footer_text: "所有业务正常"
        }}
        stats={stats}
      />
    );

    expect(screen.getByRole("contentinfo")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Beehive Blog" })).toBeInTheDocument();
    expect(screen.getByText("持续整理公开文章、笔记和项目。")).toBeInTheDocument();
    expect(screen.getByText("广东 深圳")).toBeInTheDocument();
    expect(screen.getByText("湘ICP备2023015794号-2")).toBeInTheDocument();
    expect(screen.getByText("公网安备 44030000000000号")).toBeInTheDocument();
    expect(screen.getByText("文章 12")).toBeInTheDocument();
    expect(screen.getByText("笔记 4")).toBeInTheDocument();
    expect(screen.getByText("项目 3")).toBeInTheDocument();
    expect(screen.getByText("标签 16")).toBeInTheDocument();
  });

  it("omits optional website and beian rows when missing", () => {
    render(
      <AnzhiyuFooter
        author={{
          name: "Beehive",
          description: "个人博客与知识中台"
        }}
        stats={stats}
      />
    );

    expect(screen.getByRole("heading", { name: "Beehive" })).toBeInTheDocument();
    expect(screen.queryByText(/ICP备/)).not.toBeInTheDocument();
    expect(screen.queryByText("个人网站")).not.toBeInTheDocument();
  });

  it("keeps the same semantic structure under dark theme", () => {
    document.documentElement.dataset.theme = "dark";
    render(
      <AnzhiyuFooter
        author={{
          name: "Beehive",
          description: "个人博客与知识中台"
        }}
        site={{
          name: "Beehive",
          url: "",
          subtitle: "",
          description: "",
          keywords: "",
          logo_url: "",
          favicon_url: "",
          icp_beian: "",
          police_beian: "",
          footer_text: ""
        }}
        stats={stats}
      />
    );

    expect(screen.getByRole("contentinfo")).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "服务" })).toBeInTheDocument();
    expect(screen.getByLabelText("站点统计")).toBeInTheDocument();
  });
});
