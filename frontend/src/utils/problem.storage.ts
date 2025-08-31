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
