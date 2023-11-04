import { expect, it } from "vitest";
import { parseProblemInfoToml } from "./problem.info";

const infoToml = `
title = "A + B"
timelimit = 2.0
forum = "https://github.com/yosupo06/library-checker-problems/issues/32"

[[tests]]
name = "example.in"
number = 2
[[tests]]
name = "random.cpp"
number = 10

[[solutions]]
name = "wa.cpp"
wrong = true
[[solutions]]
name = "ac_func.cpp"
function = true

[params]
A_AND_B_MAX = 1_000_000_000
LONG_LONG_PARAM = 1_000_000_000_000_000_000
`;

it("parse", () => {
  const data = parseProblemInfoToml(infoToml);
  expect(data.title).toBe("A + B");
  expect(data.timeLimit).toBe(2);
  expect(data.forum).toBe(
    "https://github.com/yosupo06/library-checker-problems/issues/32"
  );
  expect(data.tests).toStrictEqual([
    {
      name: "example.in",
      number: 2,
    },
    {
      name: "random.cpp",
      number: 10,
    },
  ]);
  expect(data.params).toStrictEqual({
    A_AND_B_MAX: 1_000_000_000n,
    LONG_LONG_PARAM: 1_000_000_000_000_000_000n,
  });
});

const infoTomlWithoutParams = `
title = "A + B"
timelimit = 2.0
forum = "https://github.com/yosupo06/library-checker-problems/issues/32"

[[tests]]
name = "example.in"
number = 2
[[tests]]
name = "random.cpp"
number = 10

[[solutions]]
name = "wa.cpp"
wrong = true
[[solutions]]
name = "ac_func.cpp"
function = true
`;

it("parseWithoutParams", () => {
  const data = parseProblemInfoToml(infoTomlWithoutParams);
  expect(data.title).toBe("A + B");
  expect(data.timeLimit).toBe(2);
  expect(data.forum).toBe(
    "https://github.com/yosupo06/library-checker-problems/issues/32"
  );
  expect(data.tests).toStrictEqual([
    {
      name: "example.in",
      number: 2,
    },
    {
      name: "random.cpp",
      number: 10,
    },
  ]);
  expect(data.params).toStrictEqual({});
});
