import { expect, it } from "vitest";
import { refactorTestCase } from "./hack";

it("refactorTestCase remove CR/LF", () => {
  expect(refactorTestCase("1\r\n2\r\n")).toStrictEqual("1\n2\n");
});

it("refactorTestCase add new line at last", () => {
  expect(refactorTestCase("1 2")).toStrictEqual("1 2\n");
});

// TODO: what is the expected behavior of this case?
it("refactorTestCase(empty string)", () => {
  expect(refactorTestCase("")).toStrictEqual("\n");
});
