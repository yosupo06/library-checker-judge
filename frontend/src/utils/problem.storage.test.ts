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
  overallVersion: "overallVersion",
  testCasesVersion: "testCaseVersion",
};

it("taskURL", () => {
  expect(taskURL(baseURL, problem)).toEqual(
    new URL("v4/files/aplusb/overallVersion/aplusb/task.md", baseURL),
  );
});

it("infoTomlURL", () => {
  expect(infoURL(baseURL, problem)).toEqual(
    new URL("v4/files/aplusb/overallVersion/aplusb/info.toml", baseURL),
  );
});

it("inFileURL", () => {
  expect(inFileURL(baseURL, problem, "example_00")).toEqual(
    new URL("v4/examples/aplusb/testCaseVersion/in/example_00.in", baseURL),
  );
});

it("outFileURL", () => {
  expect(outFileURL(baseURL, problem, "example_00")).toEqual(
    new URL("v4/examples/aplusb/testCaseVersion/out/example_00.out", baseURL),
  );
});
