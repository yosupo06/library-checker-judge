// Typed REST client using openapi-fetch + openapi-typescript types
import createClient from "openapi-fetch";
import type { components, paths } from "../openapi/types";
// Use OpenAPI-generated types directly

const getRestBaseUrl = (): string => {
  const url = import.meta.env.VITE_REST_API_URL as string | undefined;
  if (!url) {
    throw new Error(
      "VITE_REST_API_URL is not set. Please configure REST API endpoint.",
    );
  }
  return url.replace(/\/$/, "");
};

const REST_BASE = getRestBaseUrl();
const client = createClient<paths>({ baseUrl: REST_BASE });

function unwrap<T>(r: { data?: T; error?: unknown; response: Response }) {
  if (r.error) {
    const status = `${r.response.status} ${r.response.statusText}`;
    const msg = typeof r.error === "string" ? r.error : JSON.stringify(r.error);
    throw new Error(`REST ${r.response.url} failed: ${status} ${msg}`);
  }
  return r.data as T;
}

export async function fetchRanking(skip = 0, limit = 100) {
  const r = await client.GET("/ranking", {
    params: { query: { skip, limit } },
  });
  return unwrap<components["schemas"]["RankingResponse"]>(r);
}

export async function fetchProblemList(): Promise<
  components["schemas"]["ProblemListResponse"]
> {
  const r = await client.GET("/problems");
  return unwrap<components["schemas"]["ProblemListResponse"]>(r);
}

export async function fetchProblemInfo(
  name: string,
): Promise<components["schemas"]["ProblemInfoResponse"]> {
  const r = await client.GET("/problems/{name}", {
    params: { path: { name } },
  });
  return unwrap<components["schemas"]["ProblemInfoResponse"]>(r);
}

export type {};
