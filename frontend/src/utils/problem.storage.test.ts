import { expect, it } from "vitest";
import {
  inFileURL,
  infoURL,
  outFileURL,
  ProblemVersion,
  taskURL,
} from "./problem.storage";

const baseURL = new URL("https://storage.googleapis.com/my-bucket/");
const problem: ProblemVersion = {
  name: "aplusb",
  version: "version",
  testCasesVersion: "testCaseVersion",
};

it("taskURL", () => {
  expect(taskURL(baseURL, problem)).toEqual(
    new URL("aplusb/files/version/task.md", baseURL),
  );
});

it("infoTomlURL", () => {
  expect(infoURL(baseURL, problem)).toEqual(
    new URL("aplusb/files/version/info.toml", baseURL),
  );
});

it("inFileURL", () => {
  expect(inFileURL(baseURL, problem, "example_00")).toEqual(
    new URL("aplusb/testcase/testCaseVersion/in/example_00.in", baseURL),
  );
});

it("outFileURL", () => {
  expect(outFileURL(baseURL, problem, "example_00")).toEqual(
    new URL("aplusb/testcase/testCaseVersion/out/example_00.out", baseURL),
  );
});
