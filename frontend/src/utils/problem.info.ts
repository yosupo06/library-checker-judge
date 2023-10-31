import { JsonMap, parse } from "@iarna/toml";

export type ProblemInfoToml = {
  title?: string;
  timeLimit?: number;
  forum?: string;

  tests: {
    name: string;
    number: number;
  }[];
  params: { [key: string]: number };
};

export const parseProblemInfoToml = (toml: string): ProblemInfoToml => {
  const infoJsonMap = parse(toml);

  const tests = (() => {
    const data = infoJsonMap["tests"];
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
    const data = infoJsonMap["params"];
    const params: { [key: string]: number } = {};
    Object.entries(data).forEach(([key, value]) => {
      if (typeof value === "number") {
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

const readString = (data: JsonMap, key: string): string | undefined => {
  if (!(key in data)) return undefined;
  const v = data[key];
  if (typeof v !== "string") {
    return undefined;
  }
  return v;
};

const readNumber = (data: JsonMap, key: string): number | undefined => {
  if (!(key in data)) return undefined;
  const v = data[key];
  if (typeof v !== "number") {
    return undefined;
  }
  return v;
};
