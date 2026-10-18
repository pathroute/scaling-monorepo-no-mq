import type { Flag, Issue, IssueStatus, Message, Todo } from "./types.js";

/** An error response from a service. */
export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export type FetchFn = (input: string, init?: RequestInit) => Promise<Response>;

/**
 * Typed client for the suite's services. `baseUrl` is the ingress that routes
 * /todos, /issues, /messages and /admin to the right service.
 */
export class ApiClient {
  private readonly baseUrl: string;
  private readonly token: string;
  private readonly fetchFn: FetchFn;

  constructor(baseUrl: string, token: string, fetchFn: FetchFn = (i, init) => fetch(i, init)) {
    this.baseUrl = baseUrl.replace(/\/+$/, "");
    this.token = token;
    this.fetchFn = fetchFn;
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = { Authorization: `Bearer ${this.token}` };
    if (body !== undefined) headers["Content-Type"] = "application/json";
    const res = await this.fetchFn(this.baseUrl + path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    if (!res.ok) {
      let message = res.statusText;
      try {
        message = ((await res.json()) as { error?: string }).error ?? message;
      } catch {
        // Not JSON; keep the status text.
      }
      throw new ApiError(res.status, message);
    }
    return res.status === 204 ? (undefined as T) : ((await res.json()) as T);
  }

  listTodos(): Promise<Todo[]> {
    return this.request("GET", "/todos");
  }

  createTodo(title: string): Promise<Todo> {
    return this.request("POST", "/todos", { title });
  }

  completeTodo(id: string): Promise<Todo> {
    return this.request("POST", `/todos/${encodeURIComponent(id)}/done`);
  }

  listIssues(status?: IssueStatus): Promise<Issue[]> {
    return this.request("GET", status ? `/issues?status=${status}` : "/issues");
  }

  createIssue(title: string, body = ""): Promise<Issue> {
    return this.request("POST", "/issues", { title, body });
  }

  transitionIssue(id: string, status: IssueStatus): Promise<Issue> {
    return this.request("POST", `/issues/${encodeURIComponent(id)}/status`, { status });
  }

  listMessages(): Promise<Message[]> {
    return this.request("GET", "/messages");
  }

  listFlags(): Promise<Flag[]> {
    return this.request("GET", "/admin/flags");
  }

  setFlag(name: string, enabled: boolean): Promise<Flag> {
    return this.request("PUT", `/admin/flags/${encodeURIComponent(name)}`, { enabled });
  }
}
