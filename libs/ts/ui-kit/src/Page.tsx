import type { ReactNode } from "react";

export type PageProps = {
  title: string;
  error?: string | null;
  children: ReactNode;
};

/** Page shell: title, optional error banner, content. */
export function Page({ title, error, children }: PageProps) {
  return (
    <main className="sk-page">
      <h1>{title}</h1>
      {error ? <div role="alert">{error}</div> : null}
      {children}
    </main>
  );
}
