export type ProblemVersion = {
  name: string;
  version: string;
  overallVersion: string;
  testCasesVersion: string;
};

export const taskURL = (baseURL: URL, problem: ProblemVersion) => {
  // v4: files/{problem}/{overall_version}/{problem}/task.md
  return new URL(
    `files/${problem.name}/${problem.overallVersion}/${problem.name}/task.md`,
    baseURL,
  );
};

export const infoURL = (baseURL: URL, problem: ProblemVersion) => {
  // v4: files/{problem}/{overall_version}/{problem}/info.toml
  return new URL(
    `files/${problem.name}/${problem.overallVersion}/${problem.name}/info.toml`,
    baseURL,
  );
};

export const solveHppURL = (baseURL: URL, problem: ProblemVersion) => {
  // v4: files/{problem}/{overall_version}/{problem}/grader/solve.hpp
  return new URL(
    `files/${problem.name}/${problem.overallVersion}/${problem.name}/grader/solve.hpp`,
    baseURL,
  );
};

export const inFileURL = (
  baseURL: URL,
  problem: ProblemVersion,
  name: string,
) => {
  // v4: examples/{problem}/{testcase_hash}/in/{name}.in
  return new URL(
    `examples/${problem.name}/${problem.testCasesVersion}/in/${name}.in`,
    baseURL,
  );
};

export const outFileURL = (
  baseURL: URL,
  problem: ProblemVersion,
  name: string,
) => {
  // v4: examples/{problem}/{testcase_hash}/out/{name}.out
  return new URL(
    `examples/${problem.name}/${problem.testCasesVersion}/out/${name}.out`,
    baseURL,
  );
};
