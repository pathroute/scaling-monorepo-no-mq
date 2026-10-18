import { describe, expect, jest, test } from "@jest/globals";
import type { Message } from "@suite/api-client";
import { render, screen } from "@testing-library/react";

import { App, openRate } from "./App.js";

const msg = (id: string, opens: number, category?: string): Message => ({
  id,
  org: "acme",
  to: `${id}@example.com`,
  subject: `subject ${id}`,
  body: "",
  opens,
  category,
  sent_at: "",
});

describe("mail-web", () => {
  test("openRate", () => {
    expect(openRate([])).toBe(0);
    expect(openRate([msg("a", 0), msg("b", 2), msg("c", 1)])).toBe(67);
  });

  test("lists messages by opens", async () => {
    render(<App client={{ listMessages: jest.fn(async () => [msg("a", 0), msg("b", 3, "sales")]) }} />);
    expect(await screen.findByText("Open rate: 50%")).toBeInTheDocument();
    const rows = screen.getAllByRole("row").slice(1);
    expect(rows[0]).toHaveTextContent("b@example.com");
    expect(rows[0]).toHaveTextContent("sales");
    expect(rows[1]).toHaveTextContent("unclassified");
  });
});
