import React from "react";
import { Alert, Box, CircularProgress } from "@mui/material";
import { Lang } from "../contexts/LangContext";
import { StatementData, useStatementParser } from "../utils/statement.parser";
import KatexRender from "../components/katex/KatexRender";
import { useQueries, useQuery } from "@tanstack/react-query";
import { ProblemInfoToml, parseProblemInfoToml } from "../utils/problem.info";

const Statement: React.FC<{
  lang: Lang;
  data: StatementData;
}> = (props) => {
  const statement = useStatementParser(props.lang, props.data);

  if (statement.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (statement.isError) {
    return (
      <>
        <Box>
          <Alert severity="error">{(statement.error as Error).message}</Alert>
        </Box>
      </>
    );
  }

  return (
    <Box>
      <KatexRender text={statement.data} />
    </Box>
  );
};

export default Statement;

export const useProblemInfoTomlQuery = (baseUrl: URL) => {
  const infoTomlQuery = useQuery(
    ["statement", baseUrl.href, "info.toml"],
    async () =>
      fetch(new URL("info.toml", baseUrl.href)).then((r) => {
        if (r.status == 200) {
          return r.text();
        } else {
          return Promise.reject("failed to fetch info.toml:" + r.status);
        }
      }),
  );

  return useQuery({
    queryKey: ["statement", baseUrl.href, "parse-info"],
    queryFn: () => parseProblemInfoToml(infoTomlQuery.data ?? ""),
    enabled: infoTomlQuery.isSuccess,
  });
};

export const useStatement = (baseUrl: URL) => {
  return useQuery(["statement", baseUrl.href, "task.md"], () =>
    fetch(new URL("task.md", baseUrl.href)).then((r) => {
      if (r.status == 200) {
        return r.text();
      } else {
        return Promise.reject("failed to fetch task.md:" + r.status);
      }
    }),
  );
};

export const useSolveHpp = (baseUrl: URL) => {
  return useQuery(["statement", baseUrl.href, "solve.hpp"], () =>
    fetch(new URL("grader/solve.hpp", baseUrl.href)).then((r) => {
      if (r.status == 200) {
        return r.text();
      } else if (r.status == 404) {
        return null;
      }
    }),
  );
};

export const useExamples = (info: ProblemInfoToml, baseUrl: URL) => {
  const exampleNumber = (() => {
    return info.tests.find((v) => v.name === "example.in")?.number ?? 0;
  })();

  const examples = Array.from(Array(exampleNumber), (_, k) => `example_0${k}`);
  const inExampleQueries = useQueries({
    queries: examples.map((name) => {
      const inName = `in/${name}.in`;
      return {
        queryKey: [baseUrl.href, inName],
        queryFn: () =>
          fetch(new URL(inName, baseUrl.href)).then((r) => {
            if (r.status == 200) {
              return r.text();
            } else {
              return Promise.reject("failed to fetch task.md:" + r.status);
            }
          }),
      };
    }),
  });
  const outExampleQueries = useQueries({
    queries: examples.map((name) => {
      const outName = `out/${name}.out`;
      return {
        queryKey: [baseUrl.href, outName],
        queryFn: () =>
          fetch(new URL(outName, baseUrl.href)).then((r) => {
            if (r.status == 200) {
              return r.text();
            } else {
              return Promise.reject("failed to fetch task.md:" + r.status);
            }
          }),
      };
    }),
  });

  const examplesDict: { [name: string]: string } = {};
  examples.forEach((name, index) => {
    const query = inExampleQueries[index];
    if (query.isSuccess) {
      examplesDict[`${name}.in`] = query.data;
    }
  });
  examples.forEach((name, index) => {
    const query = outExampleQueries[index];
    if (query.isSuccess) {
      examplesDict[`${name}.out`] = query.data;
    }
  });

  return examplesDict;
};
