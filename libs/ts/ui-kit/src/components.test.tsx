import { describe, expect, jest, test } from "@jest/globals";
import { render, screen } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";
import { useState } from "react";

import { DataTable, Page, TextField } from "./index.js";

describe("TextField", () => {
  test("is labelled and reports changes", async () => {
    const onChange = jest.fn();
    function Wrapper() {
      const [v, setV] = useState("");
      return (
        <TextField
          label="Title"
          value={v}
          onChange={(x) => {
            setV(x);
            onChange(x);
          }}
        />
      );
    }
    render(<Wrapper />);
    await userEvent.type(screen.getByLabelText("Title"), "hi");
    expect(onChange).toHaveBeenLastCalledWith("hi");
    expect(screen.getByLabelText("Title")).toHaveValue("hi");
  });
});

describe("DataTable", () => {
  const columns = [{ header: "Name", cell: (r: { id: string; name: string }) => r.name }];

  test("renders rows", () => {
    render(<DataTable columns={columns} rows={[{ id: "1", name: "Ann" }]} rowKey={(r) => r.id} />);
    expect(screen.getByRole("columnheader", { name: "Name" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "Ann" })).toBeInTheDocument();
  });

  test("shows the empty message", () => {
    render(<DataTable columns={columns} rows={[]} rowKey={(r) => r.id} empty="No people" />);
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.getByText("No people")).toBeInTheDocument();
  });
});

describe("Page", () => {
  test("shows title and error", () => {
    render(
      <Page title="Todos" error="boom">
        body
      </Page>,
    );
    expect(screen.getByRole("heading", { name: "Todos" })).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("boom");
  });
});
