import { AnyJson, JsonMap, parse } from "@iarna/toml";

export type ProblemInfoToml = {
  title?: string;
  timeLimit?: number;
  forum?: string;

  tests: {
    name: string;
    number: number;
  }[];
  params: { [key: string]: bigint };
};

export const parseProblemInfoToml = (toml: string): ProblemInfoToml => {
  const infoJsonMap = parse(toml);

  const tests = (() => {
    const data = readField(infoJsonMap, "tests");
    if (!data) return [];
    if (!Array.isArray(data)) return [];
    const tests: {
      name: string;
      number: number;
    }[] = [];
    data.map((e) => {
      {
        if (
          typeof e === "object" &&
          !(e instanceof Array) &&
          !(e instanceof Date)
        ) {
          const name = readString(e, "name");
          const number = readNumber(e, "number");

          if (name && number) {
            tests.push({ name: name, number: number });
          }
        }
      }
    });
    return tests;
  })();

  const params = (() => {
    const data = readField(infoJsonMap, "params");
    if (!data) return {};
    const params: { [key: string]: bigint } = {};
    Object.entries(data).forEach(([key, value]) => {
      if (typeof value === "number") {
        params[key] = BigInt(value);
      }
      if (typeof value === "bigint") {
        params[key] = value;
      }
    });
    return params;
  })();

  return {
    title: readString(infoJsonMap, "title"),
    timeLimit: readNumber(infoJsonMap, "timelimit"),
    forum: readString(infoJsonMap, "forum"),
    tests: tests,
    params: params,
  };
};

const readField = (data: JsonMap, key: string): AnyJson | undefined => {
  if (!(key in data)) return undefined;
  return data[key];
};

const readString = (data: JsonMap, key: string): string | undefined => {
  const v = readField(data, key);
  if (typeof v !== "string") {
    return undefined;
  }
  return v;
};

const readNumber = (data: JsonMap, key: string): number | undefined => {
  const v = readField(data, key);
  if (typeof v !== "number") {
    return undefined;
  }
  return v;
};
