import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LoginPage } from "./LoginPage";

describe("LoginPage", () => {
  it("submits username and password to handler", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);

    render(<LoginPage isSubmitting={false} error="" onSubmit={onSubmit} />);

    await user.type(screen.getByPlaceholderText("请输入管理员用户名"), "admin");
    await user.type(screen.getByPlaceholderText("请输入管理员密码"), "change-me");
    await user.click(screen.getByRole("button", { name: "进入后台" }));

    expect(onSubmit).toHaveBeenCalledWith("admin", "change-me");
  });
});
