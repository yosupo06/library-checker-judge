export type ProblemVersion = {
  name: string;
  version: string;
  testCasesVersion: string;
};

export const taskURL = (baseURL: URL, problem: ProblemVersion) => {
  return new URL(`${problem.name}/files/${problem.version}/task.md`, baseURL);
};

export const infoURL = (baseURL: URL, problem: ProblemVersion) => {
  return new URL(`${problem.name}/files/${problem.version}/info.toml`, baseURL);
};

export const solveHppURL = (baseURL: URL, problem: ProblemVersion) => {
  return new URL(
    `${problem.name}/files/${problem.version}/grader/solve.hpp`,
    baseURL,
  );
};

export const inFileURL = (
  baseURL: URL,
  problem: ProblemVersion,
  name: string,
) => {
  return new URL(
    `${problem.name}/testcase/${problem.testCasesVersion}/in/${name}.in`,
    baseURL,
  );
};

export const outFileURL = (
  baseURL: URL,
  problem: ProblemVersion,
  name: string,
) => {
  return new URL(
    `${problem.name}/testcase/${problem.testCasesVersion}/out/${name}.out`,
    baseURL,
  );
};
