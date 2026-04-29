import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LinksPage } from "./LinksPage";
import * as api from "../lib/api";
import type { AuthSession, Link, LinkAnalytics } from "../types";

const session: AuthSession = {
  authenticated: true,
  username: "admin",
};

const links: Link[] = [
  {
    id: 1,
    code: "demo",
    target_url: "https://example.com/demo",
    remark: "演示链接",
    tags: ["demo", "docs"],
    enabled: true,
    click_count: 12,
  },
];

const analytics: LinkAnalytics = {
  link: links[0],
  range_days: 7,
  recent_clicks: 8,
  unique_ips: 3,
  last_visited_at: "2026-04-12T10:15:00Z",
  time_series: [
    { bucket: "2026-04-10", clicks: 2 },
    { bucket: "2026-04-11", clicks: 3 },
    { bucket: "2026-04-12", clicks: 3 },
  ],
  top_referrers: [{ name: "mp.weixin.qq.com", count: 5 }],
  top_clients: [{ name: "微信", count: 4 }],
  recent_visits: [
    {
      visited_at: "2026-04-12T10:15:00Z",
      ip_masked: "10.0.0.*",
      referer: "https://mp.weixin.qq.com/s/demo",
      referer_host: "mp.weixin.qq.com",
      user_agent: "Mozilla/5.0 MicroMessenger",
      client_name: "微信",
      client_type: "app",
      device_type: "mobile",
      os: "iOS",
    },
  ],
};

describe("LinksPage", () => {
  it("loads analytics panel", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "getLinkAnalytics").mockResolvedValue(analytics);

    render(
      <LinksPage
        session={session}
        links={links}
        isLoading={false}
        error=""
        onReload={vi.fn().mockResolvedValue(undefined)}
        onLogout={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "分析" }));

    await waitFor(() => {
      expect(screen.getByText("demo 的访问分析")).toBeInTheDocument();
      expect(screen.getAllByText("mp.weixin.qq.com").length).toBeGreaterThan(0);
      expect(screen.getByText("10.0.0.*")).toBeInTheDocument();
    });
  });

  it("renders link list", async () => {
    render(
      <LinksPage
        session={session}
        links={links}
        isLoading={false}
        error=""
        onReload={vi.fn().mockResolvedValue(undefined)}
        onLogout={vi.fn()}
      />,
    );

    await waitFor(() => {
      expect(screen.getByText("demo")).toBeInTheDocument();
      expect(screen.getByText("https://example.com/demo")).toBeInTheDocument();
      expect(screen.getByText("演示链接")).toBeInTheDocument();
    });
  });

  it("copies short link when copy icon is clicked", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    });

    render(
      <LinksPage
        session={session}
        links={links}
        isLoading={false}
        error=""
        onReload={vi.fn().mockResolvedValue(undefined)}
        onLogout={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "复制 demo 短链" }));

    expect(writeText).toHaveBeenCalledWith("http://localhost:3000/demo");
  });

  it("focuses search with slash shortcut and filters links", async () => {
    const user = userEvent.setup();

    render(
      <LinksPage
        session={session}
        links={[
          ...links,
          {
            id: 2,
            code: "other",
            target_url: "https://example.com/other",
            remark: "其他链接",
            tags: ["marketing"],
            enabled: true,
            click_count: 1,
          },
        ]}
        isLoading={false}
        error=""
        onReload={vi.fn().mockResolvedValue(undefined)}
        onLogout={vi.fn()}
      />,
    );

    await user.keyboard("/");
    expect(screen.getByRole("textbox", { name: "搜索短链" })).toHaveFocus();

    await user.keyboard("other");

    await waitFor(() => {
      expect(screen.getByRole("link", { name: "other" })).toBeInTheDocument();
      expect(screen.queryByRole("link", { name: "demo" })).not.toBeInTheDocument();
    });
  });

  it("opens edit modal", async () => {
    const user = userEvent.setup();

    render(
      <LinksPage
        session={session}
        links={links}
        isLoading={false}
        error=""
        onReload={vi.fn().mockResolvedValue(undefined)}
        onLogout={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "编辑" }));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(screen.getByDisplayValue("demo")).toBeInTheDocument();
    });
  });

  it("filters by tag when tag chip is clicked", async () => {
    const user = userEvent.setup();

    render(
      <LinksPage
        session={session}
        links={[
          ...links,
          {
            id: 2,
            code: "other",
            target_url: "https://example.com/other",
            remark: "其他链接",
            tags: ["marketing"],
            enabled: true,
            click_count: 1,
          },
        ]}
        isLoading={false}
        error=""
        onReload={vi.fn().mockResolvedValue(undefined)}
        onLogout={vi.fn()}
      />,
    );

    await user.click(screen.getAllByRole("button", { name: "#demo" })[0]);

    await waitFor(() => {
      expect(screen.getByRole("link", { name: "demo" })).toBeInTheDocument();
      expect(screen.queryByRole("link", { name: "other" })).not.toBeInTheDocument();
    });
  });
});
