import type { ApiClient, Todo } from "@suite/api-client";
import { Button, DataTable, Page, TextField } from "@suite/ui-kit";
import { type FormEvent, useEffect, useState } from "react";

export type TodoClient = Pick<ApiClient, "listTodos" | "createTodo" | "completeTodo">;

export function App({ client }: { client: TodoClient }) {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [title, setTitle] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    client.listTodos().then(setTodos, (e: Error) => setError(e.message));
  }, [client]);

  async function add(e: FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    try {
      const todo = await client.createTodo(title.trim());
      setTodos((ts) => [...ts, todo]);
      setTitle("");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function complete(id: string) {
    try {
      const done = await client.completeTodo(id);
      setTodos((ts) => ts.map((t) => (t.id === id ? done : t)));
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <Page title="Todos" error={error}>
      <form onSubmit={add}>
        <TextField label="New todo" value={title} onChange={setTitle} />
        <Button type="submit">Add</Button>
      </form>
      <DataTable
        rows={todos}
        rowKey={(t) => t.id}
        empty="No todos yet."
        columns={[
          { header: "Title", cell: (t) => t.title },
          {
            header: "Status",
            cell: (t) =>
              t.done ? "done" : (
                <Button variant="secondary" onClick={() => complete(t.id)}>
                  Mark done
                </Button>
              ),
          },
        ]}
      />
    </Page>
  );
}
