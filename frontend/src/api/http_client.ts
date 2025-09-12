// Lightweight HTTP client for REST endpoints migrated from gRPC-Web

const getRestBaseUrl = (): string => {
  const explicit = import.meta.env.VITE_REST_API_URL;
  if (explicit) return explicit.replace(/\/$/, "");

  // Fallback: try to derive from gRPC base URL if present
  const api = import.meta.env.VITE_API_URL;
  if (api) {
    try {
      const u = new URL(api);
      // If port is 12380 (gRPC-web), use 12381 (REST)
      if (u.port === "12380" || u.port === "") {
        u.port = "12381";
      }
      return u.origin;
    } catch {
      // ignore
    }
  }
  // Local default for dev
  return "http://localhost:12381";
};

const REST_BASE = getRestBaseUrl();

export async function fetchRanking(skip = 0, limit = 100) {
  const url = new URL("/ranking", REST_BASE);
  url.searchParams.set("skip", String(skip));
  url.searchParams.set("limit", String(limit));
  const res = await fetch(url.toString(), {
    method: "GET",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(
      `REST /ranking failed: ${res.status} ${res.statusText} ${text}`,
    );
  }
  return res.json();
}

export type {};
