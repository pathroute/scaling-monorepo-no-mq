/**
 * @jest-environment node
 */
import { describe, expect, jest, test } from "@jest/globals";

import { ApiClient, ApiError, type FetchFn } from "./index.js";

function fakeFetch(status: number, body: unknown) {
  return jest.fn<FetchFn>(async () => new Response(body === undefined ? null : JSON.stringify(body), { status }));
}

describe("ApiClient", () => {
  test("sends the token and JSON body", async () => {
    const fetchFn = fakeFetch(201, { id: "t1", title: "x" });
    const client = new ApiClient("https://api.example/", "tok", fetchFn);
    const todo = await client.createTodo("x");
    expect(todo.id).toBe("t1");
    const [url, init] = fetchFn.mock.calls[0]!;
    expect(url).toBe("https://api.example/todos");
    expect(init?.method).toBe("POST");
    expect(init?.body).toBe(JSON.stringify({ title: "x" }));
    expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer tok");
  });

  test("builds query strings and escapes ids", async () => {
    const fetchFn = fakeFetch(200, []);
    const client = new ApiClient("https://api.example", "tok", fetchFn);
    await client.listIssues("closed");
    await client.setFlag("a/b", true);
    expect(fetchFn.mock.calls.map((c) => c[0])).toEqual([
      "https://api.example/issues?status=closed",
      "https://api.example/admin/flags/a%2Fb",
    ]);
  });

  test("turns error responses into ApiError", async () => {
    const client = new ApiClient("https://api.example", "tok", fakeFetch(403, { error: "missing role admin" }));
    await expect(client.listFlags()).rejects.toEqual(new ApiError(403, "missing role admin"));
  });
});
