// Wire types; keep in sync with the Go services' api packages (services/*/api).

export type Todo = {
  id: string;
  org: string;
  title: string;
  done: boolean;
  created_at: string;
};

export type IssueStatus = "open" | "in_progress" | "closed";

export type Issue = {
  id: string;
  org: string;
  title: string;
  body: string;
  status: IssueStatus;
  created_at: string;
};

export type Message = {
  id: string;
  org: string;
  to: string;
  subject: string;
  body: string;
  opens: number;
  category?: string;
  sent_at: string;
};

export type Flag = {
  name: string;
  enabled: boolean;
};
