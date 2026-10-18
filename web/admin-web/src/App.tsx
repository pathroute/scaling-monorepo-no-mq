import type { ApiClient, Flag } from "@suite/api-client";
import { Button, DataTable, Page } from "@suite/ui-kit";
import { useEffect, useState } from "react";

export type AdminClient = Pick<ApiClient, "listFlags" | "setFlag">;

export function App({ client }: { client: AdminClient }) {
  const [flags, setFlags] = useState<Flag[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    client.listFlags().then(setFlags, (e: Error) => setError(e.message));
  }, [client]);

  async function toggle(f: Flag) {
    try {
      const updated = await client.setFlag(f.name, !f.enabled);
      setFlags((fs) => fs.map((x) => (x.name === f.name ? updated : x)));
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <Page title="Feature flags" error={error}>
      <DataTable
        rows={flags}
        rowKey={(f) => f.name}
        empty="No flags."
        columns={[
          { header: "Flag", cell: (f) => f.name },
          { header: "State", cell: (f) => (f.enabled ? "on" : "off") },
          {
            header: "",
            cell: (f) => (
              <Button variant="secondary" onClick={() => toggle(f)}>
                {f.enabled ? "Disable" : "Enable"} {f.name}
              </Button>
            ),
          },
        ]}
      />
    </Page>
  );
}
