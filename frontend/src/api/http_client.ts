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

function authHeaders(
  idToken?: string | null,
): Record<string, string> | undefined {
  if (!idToken) return undefined;
  return { Authorization: `Bearer ${idToken}` };
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

export async function fetchLangList(): Promise<
  components["schemas"]["LangListResponse"]
> {
  const r = await client.GET("/langs");
  return unwrap<components["schemas"]["LangListResponse"]>(r);
}

export async function fetchProblemCategories(): Promise<
  components["schemas"]["ProblemCategoriesResponse"]
> {
  const r = await client.GET("/categories");
  return unwrap<components["schemas"]["ProblemCategoriesResponse"]>(r);
}

// Auth
export async function fetchCurrentUserInfo(
  idToken?: string | null,
): Promise<components["schemas"]["CurrentUserInfoResponse"]> {
  const r = await client.GET("/auth/current_user", {
    headers: authHeaders(idToken),
  });
  return unwrap<components["schemas"]["CurrentUserInfoResponse"]>(r);
}

export async function registerUser(
  name: string,
  idToken?: string | null,
): Promise<components["schemas"]["RegisterResponse"]> {
  const r = await client.POST("/auth/register", {
    body: { name },
    headers: authHeaders(idToken),
  });
  return unwrap<components["schemas"]["RegisterResponse"]>(r);
}

export async function patchCurrentUserInfo(
  user: components["schemas"]["User"],
  idToken?: string | null,
): Promise<components["schemas"]["ChangeCurrentUserInfoResponse"]> {
  const r = await client.PATCH("/auth/current_user", {
    body: { user },
    headers: authHeaders(idToken),
  });
  return unwrap<components["schemas"]["ChangeCurrentUserInfoResponse"]>(r);
}

export async function fetchUserInfo(
  name: string,
): Promise<components["schemas"]["UserInfoResponse"]> {
  const r = await client.GET("/users/{name}", { params: { path: { name } } });
  return unwrap<components["schemas"]["UserInfoResponse"]>(r);
}

export async function patchUserInfo(
  name: string,
  user: components["schemas"]["User"],
  idToken?: string | null,
): Promise<components["schemas"]["ChangeUserInfoResponse"]> {
  const r = await client.PATCH("/users/{name}", {
    params: { path: { name } },
    body: { user },
    headers: authHeaders(idToken),
  });
  return unwrap<components["schemas"]["ChangeUserInfoResponse"]>(r);
}

// Submissions
export type SubmitPayload = {
  problem: string;
  source: string;
  lang: string;
  tleKnockout?: boolean;
};

export type SubmissionListQuery = {
  problem?: string;
  user?: string;
  dedupUser?: boolean;
  status?: string;
  lang?: string;
  order?: string;
  skip?: number;
  limit?: number;
};

type SubmissionListQueryParams = NonNullable<
  paths["/submissions"]["get"]["parameters"]["query"]
>;

const dropUndefined = <T extends Record<string, unknown>>(obj: T): T => {
  return Object.fromEntries(
    Object.entries(obj).filter(([, value]) => value !== undefined),
  ) as T;
};

export async function postSubmit(
  payload: SubmitPayload,
  idToken?: string | null,
): Promise<components["schemas"]["SubmitResponse"]> {
  const body = dropUndefined<components["schemas"]["SubmitRequest"]>({
    problem: payload.problem,
    source: payload.source,
    lang: payload.lang,
    tle_knockout: payload.tleKnockout,
  });
  const r = await client.POST("/submit", {
    body,
    headers: authHeaders(idToken),
  });
  return unwrap<components["schemas"]["SubmitResponse"]>(r);
}

export async function fetchSubmissionList(
  query: SubmissionListQuery,
): Promise<components["schemas"]["SubmissionListResponse"]> {
  const params = dropUndefined<SubmissionListQueryParams>({
    skip: query.skip,
    limit: query.limit,
    problem: query.problem,
    status: query.status,
    user: query.user,
    dedupUser: query.dedupUser,
    lang: query.lang,
    order: query.order,
  });
  const r = await client.GET("/submissions", {
    params: { query: params },
  });
  return unwrap<components["schemas"]["SubmissionListResponse"]>(r);
}

export async function fetchSubmissionInfo(
  id: number,
): Promise<components["schemas"]["SubmissionInfoResponse"]> {
  const r = await client.GET("/submissions/{id}", {
    params: { path: { id } },
  });
  return unwrap<components["schemas"]["SubmissionInfoResponse"]>(r);
}

export type {};
