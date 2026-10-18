import { describe, expect, jest, test } from "@jest/globals";
import { render, screen } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";

import { type AdminClient, App } from "./App.js";

describe("admin-web", () => {
  test("toggles a flag", async () => {
    const client: AdminClient = {
      listFlags: jest.fn(async () => [{ name: "new_ui", enabled: false }]),
      setFlag: jest.fn(async (name: string, enabled: boolean) => ({ name, enabled })),
    };
    render(<App client={client} />);
    await userEvent.click(await screen.findByRole("button", { name: "Enable new_ui" }));
    expect(client.setFlag).toHaveBeenCalledWith("new_ui", true);
    expect(await screen.findByText("on")).toBeInTheDocument();
  });

  test("shows permission errors from toggling", async () => {
    const client: AdminClient = {
      listFlags: jest.fn(async () => [{ name: "new_ui", enabled: true }]),
      setFlag: jest.fn(async () => {
        throw new Error("missing role admin");
      }),
    };
    render(<App client={client} />);
    await userEvent.click(await screen.findByRole("button", { name: "Disable new_ui" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("missing role admin");
  });
});
