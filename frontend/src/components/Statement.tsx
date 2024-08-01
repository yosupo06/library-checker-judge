import React from "react";
import { Alert, Box, CircularProgress } from "@mui/material";
import { Lang } from "../contexts/LangContext";
import { StatementData, useStatementParser } from "../utils/statement.parser";
import KatexRender from "../components/katex/KatexRender";
import { useQueries, useQuery } from "@tanstack/react-query";
import { ProblemInfoToml, parseProblemInfoToml } from "../utils/problem.info";
import {
  ProblemVersion,
  inFileURL,
  infoURL,
  outFileURL,
  solveHppURL,
  taskURL,
} from "../utils/problem.storage";

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

export const useProblemInfoTomlQuery = (
  baseURL: URL,
  problemVersion: ProblemVersion,
) => {
  const url = infoURL(baseURL, problemVersion);
  const infoTomlQuery = useQuery(
    ["statement", problemVersion, "info.toml"],
    async () =>
      fetch(url).then((r) => {
        if (r.status == 200) {
          return r.text();
        } else {
          return Promise.reject("failed to fetch info.toml:" + r.status);
        }
      }),
  );

  return useQuery({
    queryKey: ["statement", problemVersion, "parse-info"],
    queryFn: () => parseProblemInfoToml(infoTomlQuery.data ?? ""),
    enabled: infoTomlQuery.isSuccess,
  });
};

export const useStatement = (baseURL: URL, problemVersion: ProblemVersion) => {
  const url = taskURL(baseURL, problemVersion);
  return useQuery(["statement", problemVersion, "task.md"], () =>
    fetch(url).then((r) => {
      if (r.status == 200) {
        return r.text();
      } else {
        return Promise.reject("failed to fetch task.md:" + r.status);
      }
    }),
  );
};

export const useSolveHpp = (baseURL: URL, problemVersion: ProblemVersion) => {
  const url = solveHppURL(baseURL, problemVersion);
  return useQuery(["statement", problemVersion, "solve.hpp"], () =>
    fetch(url).then((r) => {
      if (r.status == 200) {
        return r.text();
      } else if (r.status == 404) {
        return null;
      }
    }),
  );
};

export const useExamples = (
  info: ProblemInfoToml,
  baseURL: URL,
  problemVersion: ProblemVersion,
) => {
  const exampleNumber = (() => {
    return info.tests.find((v) => v.name === "example.in")?.number ?? 0;
  })();

  const examples = Array.from(Array(exampleNumber), (_, k) => `example_0${k}`);
  const inExampleQueries = useQueries({
    queries: examples.map((name) => {
      const url = inFileURL(baseURL, problemVersion, name);
      return {
        queryKey: [problemVersion, "in", name],
        queryFn: () =>
          fetch(url).then((r) => {
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
      const url = outFileURL(baseURL, problemVersion, name);
      return {
        queryKey: [problemVersion, "out", name],
        queryFn: () =>
          fetch(url).then((r) => {
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
