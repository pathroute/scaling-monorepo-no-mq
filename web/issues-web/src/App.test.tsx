import { describe, expect, jest, test } from "@jest/globals";
import type { Issue, IssueStatus } from "@suite/api-client";
import { render, screen, within } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";

import { App, type IssuesClient } from "./App.js";

const issue = (id: string, title: string, status: IssueStatus = "open"): Issue => ({
  id,
  org: "acme",
  title,
  body: "",
  status,
  created_at: "",
});

describe("issues-web", () => {
  test("filters by status and closes issues", async () => {
    const all = [issue("1", "crash"), issue("2", "typo", "closed")];
    const client: IssuesClient = {
      listIssues: jest.fn(async (status?: IssueStatus) => all.filter((i) => !status || i.status === status)),
      createIssue: jest.fn(async (title: string) => issue("3", title)),
      transitionIssue: jest.fn(async (id: string, status: IssueStatus) => ({ ...all.find((i) => i.id === id)!, status })),
    };
    render(<App client={client} />);
    expect(await screen.findByText("typo")).toBeInTheDocument();

    await userEvent.selectOptions(screen.getByLabelText("Status"), "Open");
    expect(client.listIssues).toHaveBeenLastCalledWith("open");
    expect(await screen.findByText("crash")).toBeInTheDocument();
    expect(screen.queryByText("typo")).not.toBeInTheDocument();

    const row = screen.getByText("crash").closest("tr")!;
    await userEvent.click(within(row).getByRole("button", { name: "Close" }));
    expect(client.transitionIssue).toHaveBeenCalledWith("1", "closed");
    expect(await screen.findByText("No issues.")).toBeInTheDocument();
  });
});
