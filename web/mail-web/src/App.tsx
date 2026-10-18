import type { ApiClient, Message } from "@suite/api-client";
import { DataTable, Page } from "@suite/ui-kit";
import { useEffect, useState } from "react";

export type MailClient = Pick<ApiClient, "listMessages">;

/** Share of messages opened at least once, as a whole percentage. */
export function openRate(messages: Message[]): number {
  if (messages.length === 0) return 0;
  return Math.round((100 * messages.filter((m) => m.opens > 0).length) / messages.length);
}

export function App({ client }: { client: MailClient }) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    client.listMessages().then(
      (ms) => setMessages([...ms].sort((a, b) => b.opens - a.opens)),
      (e: Error) => setError(e.message),
    );
  }, [client]);

  return (
    <Page title="Mail" error={error}>
      <p>Open rate: {openRate(messages)}%</p>
      <DataTable
        rows={messages}
        rowKey={(m) => m.id}
        empty="No tracked mail."
        columns={[
          { header: "To", cell: (m) => m.to },
          { header: "Subject", cell: (m) => m.subject },
          { header: "Opens", cell: (m) => m.opens },
          { header: "Category", cell: (m) => m.category ?? "unclassified" },
        ]}
      />
    </Page>
  );
}
