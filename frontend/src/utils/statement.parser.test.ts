import { expect, it } from "vitest";
import { paramToStr } from "./statement.parser";

it("paramToStr", () => {
  expect(paramToStr(0n)).toBe("0");
  expect(paramToStr(1n)).toBe("1");

  expect(paramToStr(100_000n)).toBe("10^{5}");
  expect(paramToStr(200_000n)).toBe("2 \\times 10^{5}");
  expect(paramToStr(1_000_000n)).toBe("10^{6}");

  expect(paramToStr(1_000_000_000_000_000_000n)).toBe("10^{18}");
});
