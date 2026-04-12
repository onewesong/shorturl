import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LinksPage } from "./LinksPage";
import type { AuthSession, Link } from "../types";

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

describe("LinksPage", () => {
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
