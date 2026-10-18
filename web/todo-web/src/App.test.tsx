import { describe, expect, jest, test } from "@jest/globals";
import type { Todo } from "@suite/api-client";
import { render, screen } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";

import { App, type TodoClient } from "./App.js";

const todo = (id: string, title: string, done = false): Todo => ({ id, org: "acme", title, done, created_at: "" });

function fakeClient(initial: Todo[]): TodoClient {
  return {
    listTodos: jest.fn(async () => initial),
    createTodo: jest.fn(async (title: string) => todo("new", title)),
    completeTodo: jest.fn(async (id: string) => todo(id, "first", true)),
  };
}

describe("todo-web", () => {
  test("lists, adds and completes todos", async () => {
    const client = fakeClient([todo("1", "first")]);
    render(<App client={client} />);
    expect(await screen.findByText("first")).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText("New todo"), "second");
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(await screen.findByText("second")).toBeInTheDocument();
    expect(client.createTodo).toHaveBeenCalledWith("second");

    await userEvent.click(screen.getAllByRole("button", { name: "Mark done" })[0]!);
    expect(await screen.findByText("done")).toBeInTheDocument();
  });

  test("shows load errors", async () => {
    const client = fakeClient([]);
    client.listTodos = jest.fn(async () => {
      throw new Error("missing bearer token");
    });
    render(<App client={client} />);
    expect(await screen.findByRole("alert")).toHaveTextContent("missing bearer token");
  });
});
