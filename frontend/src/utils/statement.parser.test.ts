import { expect, it } from "vitest";
import { paramToStr } from "./statement.parser";

it("paramToStr", () => {
  expect(paramToStr(0n)).toBe("0");
  expect(paramToStr(1n)).toBe("1");
  expect(paramToStr(123456n)).toBe("123456");

  expect(paramToStr(10_000n)).toBe("10000");
  expect(paramToStr(100_000n)).toBe("10^{5}");
  expect(paramToStr(200_000n)).toBe("2 \\times 10^{5}");
  expect(paramToStr(1_000_000n)).toBe("10^{6}");
  expect(paramToStr(1_000_000_000_000_000_000n)).toBe("10^{18}");

  expect(paramToStr(512n)).toBe("512");
  expect(paramToStr(1024n)).toBe("2^{10}");
  expect(paramToStr(2048n)).toBe("2^{11}");
  expect(paramToStr(3072n)).toBe("3072");
});
