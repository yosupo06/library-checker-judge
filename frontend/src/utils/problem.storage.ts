export type ProblemVersion = {
  name: string;
  version: string;
  overallVersion: string;
  testCasesVersion: string;
};

const VERSION_PREFIX = "v4";

const withTrailingSlash = (u: URL): URL => {
  const s = u.toString();
  return new URL(s.endsWith("/") ? s : s + "/");
};

export const taskURL = (baseURL: URL, problem: ProblemVersion) => {
  // v4: v4/files/{problem}/{overall_version}/{problem}/task.md
  return new URL(
    `${VERSION_PREFIX}/files/${problem.name}/${problem.overallVersion}/${problem.name}/task.md`,
    withTrailingSlash(baseURL),
  );
};

export const infoURL = (baseURL: URL, problem: ProblemVersion) => {
  // v4: v4/files/{problem}/{overall_version}/{problem}/info.toml
  return new URL(
    `${VERSION_PREFIX}/files/${problem.name}/${problem.overallVersion}/${problem.name}/info.toml`,
    withTrailingSlash(baseURL),
  );
};

export const solveHppURL = (baseURL: URL, problem: ProblemVersion) => {
  // v4: v4/files/{problem}/{overall_version}/{problem}/grader/solve.hpp
  return new URL(
    `${VERSION_PREFIX}/files/${problem.name}/${problem.overallVersion}/${problem.name}/grader/solve.hpp`,
    withTrailingSlash(baseURL),
  );
};

export const inFileURL = (
  baseURL: URL,
  problem: ProblemVersion,
  name: string,
) => {
  // v4: v4/examples/{problem}/{testcase_hash}/in/{name}.in
  return new URL(
    `${VERSION_PREFIX}/examples/${problem.name}/${problem.testCasesVersion}/in/${name}.in`,
    withTrailingSlash(baseURL),
  );
};

export const outFileURL = (
  baseURL: URL,
  problem: ProblemVersion,
  name: string,
) => {
  // v4: v4/examples/{problem}/{testcase_hash}/out/{name}.out
  return new URL(
    `${VERSION_PREFIX}/examples/${problem.name}/${problem.testCasesVersion}/out/${name}.out`,
    withTrailingSlash(baseURL),
  );
};

// Optional convenience to create ProblemVersion consistently
// import type { ProblemInfoResponse } from "../proto/library_checker"; // keep import in caller to avoid circular deps
export const toProblemVersion = (
  problemId: string,
  info: { version: string; overallVersion: string; testcasesVersion: string },
): ProblemVersion => ({
  name: problemId,
  version: info.version,
  overallVersion: info.overallVersion,
  testCasesVersion: info.testcasesVersion,
});
import { useQuery, useQueries } from "@tanstack/react-query";
import { parseProblemInfoToml, type ProblemInfoToml } from "./problem.info";
import type { components as OpenApi } from "../openapi/types";

export type ProblemAssets = {
  info?: ProblemInfoToml;
  statement?: string;
  solveHpp?: string | null;
  examples: { [name: string]: string };
  isPending: boolean;
  error: unknown | null;
};

export const useProblemAssets = (
  baseURL: URL,
  problemId: string,
  problemInfo: OpenApi["schemas"]["ProblemInfoResponse"],
): ProblemAssets => {
  const pv = toProblemVersion(problemId, {
    version: problemInfo.version,
    overallVersion: problemInfo.overall_version,
    testcasesVersion: problemInfo.testcases_version,
  });

  const infoTomlQ = useQuery({
    queryKey: ["statement", pv, "info.toml"],
    queryFn: async () => {
      const url = infoURL(baseURL, pv);
      const r = await fetch(url);
      if (r.status === 200) return r.text();
      throw new Error("failed to fetch info.toml:" + r.status);
    },
  });

  const parsedInfoQ = useQuery({
    queryKey: ["statement", pv, "parse-info"],
    queryFn: () => parseProblemInfoToml(infoTomlQ.data ?? ""),
    structuralSharing: false,
    enabled: infoTomlQ.isSuccess,
  });

  const statementQ = useQuery({
    queryKey: ["statement", pv, "task.md"],
    queryFn: async () => {
      const url = taskURL(baseURL, pv);
      const r = await fetch(url);
      if (r.status === 200) return r.text();
      throw new Error("failed to fetch task.md:" + r.status);
    },
  });

  const solveHppQ = useQuery({
    queryKey: ["statement", pv, "solve.hpp"],
    queryFn: async () => {
      const url = solveHppURL(baseURL, pv);
      const r = await fetch(url);
      if (r.status === 200) return r.text();
      if (r.status === 404) return null;
      throw new Error("failed to fetch solve.hpp:" + r.status);
    },
  });

  const exampleNames: string[] = (() => {
    const n =
      parsedInfoQ.data?.tests.find((v) => v.name === "example.in")?.number ?? 0;
    return Array.from(Array(n), (_, k) => `example_0${k}`);
  })();

  const inQueries = useQueries({
    queries: exampleNames.map((name) => ({
      queryKey: [pv, "in", name],
      queryFn: async () => {
        const url = inFileURL(baseURL, pv, name);
        const r = await fetch(url);
        if (r.status === 200) return r.text();
        throw new Error("failed to fetch example in:" + r.status);
      },
      enabled: parsedInfoQ.isSuccess,
    })),
  });
  const outQueries = useQueries({
    queries: exampleNames.map((name) => ({
      queryKey: [pv, "out", name],
      queryFn: async () => {
        const url = outFileURL(baseURL, pv, name);
        const r = await fetch(url);
        if (r.status === 200) return r.text();
        throw new Error("failed to fetch example out:" + r.status);
      },
      enabled: parsedInfoQ.isSuccess,
    })),
  });

  const examples: { [name: string]: string } = {};
  exampleNames.forEach((name, i) => {
    if (inQueries[i]?.isSuccess)
      examples[`${name}.in`] = inQueries[i].data as string;
    if (outQueries[i]?.isSuccess)
      examples[`${name}.out`] = outQueries[i].data as string;
  });

  return {
    info: parsedInfoQ.data,
    statement: statementQ.data,
    solveHpp: solveHppQ.data,
    examples,
    isPending:
      infoTomlQ.isPending ||
      parsedInfoQ.isPending ||
      statementQ.isPending ||
      solveHppQ.isPending ||
      inQueries.some((q) => q.isPending) ||
      outQueries.some((q) => q.isPending),
    error:
      infoTomlQ.error ||
      parsedInfoQ.error ||
      statementQ.error ||
      solveHppQ.error ||
      inQueries.find((q) => q.error)?.error ||
      outQueries.find((q) => q.error)?.error ||
      null,
  };
};
