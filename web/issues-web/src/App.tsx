import type { ApiClient, Issue, IssueStatus } from "@suite/api-client";
import { Button, DataTable, Page, TextField } from "@suite/ui-kit";
import { type FormEvent, useEffect, useState } from "react";

export type IssuesClient = Pick<ApiClient, "listIssues" | "createIssue" | "transitionIssue">;

const FILTERS: Array<{ label: string; value: IssueStatus | "" }> = [
  { label: "All", value: "" },
  { label: "Open", value: "open" },
  { label: "In progress", value: "in_progress" },
  { label: "Closed", value: "closed" },
];

export function App({ client }: { client: IssuesClient }) {
  const [issues, setIssues] = useState<Issue[]>([]);
  const [filter, setFilter] = useState<IssueStatus | "">("");
  const [title, setTitle] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    client.listIssues(filter || undefined).then(setIssues, (e: Error) => setError(e.message));
  }, [client, filter]);

  async function create(e: FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    try {
      const issue = await client.createIssue(title.trim());
      if (!filter || filter === issue.status) setIssues((is) => [...is, issue]);
      setTitle("");
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function close(id: string) {
    try {
      const closed = await client.transitionIssue(id, "closed");
      setIssues((is) => (filter && filter !== "closed" ? is.filter((i) => i.id !== id) : is.map((i) => (i.id === id ? closed : i))));
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <Page title="Issues" error={error}>
      <label>
        Status{" "}
        <select value={filter} onChange={(e) => setFilter(e.target.value as IssueStatus | "")}>
          {FILTERS.map((f) => (
            <option key={f.label} value={f.value}>
              {f.label}
            </option>
          ))}
        </select>
      </label>
      <form onSubmit={create}>
        <TextField label="New issue" value={title} onChange={setTitle} />
        <Button type="submit">File issue</Button>
      </form>
      <DataTable
        rows={issues}
        rowKey={(i) => i.id}
        empty="No issues."
        columns={[
          { header: "Title", cell: (i) => i.title },
          { header: "Status", cell: (i) => i.status },
          {
            header: "",
            cell: (i) =>
              i.status === "closed" ? null : (
                <Button variant="secondary" onClick={() => close(i.id)}>
                  Close
                </Button>
              ),
          },
        ]}
      />
    </Page>
  );
}
