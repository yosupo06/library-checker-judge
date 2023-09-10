import React, { useEffect, useState } from "react";
import { Container } from "@mui/material";
import { Lang } from "../contexts/LangContext";
import { parseStatement } from "../utils/StatementParser";
import { JsonArray, JsonMap, parse } from "@iarna/toml";
import { unified } from "unified";
import remarkRehype from "remark-rehype";
import remarkParse from "remark-parse";
import rehypeStringify from "rehype-stringify";
import KatexRender from "../components/katex/KatexRender";
import { useQueries, useQuery } from "@tanstack/react-query";
import urlJoin from "url-join";

export type StatementData = {
  info: JsonMap;
  statement: string;
  examples: { [name: string]: string };
};

const Statement: React.FC<{
  lang: Lang;
  data: StatementData;
}> = (props) => {
  const [statement, setStatement] = useState("");

  useEffect(() => {
    const { info, statement, examples } = props.data;

    const rawParams = info.params || null;
    const rawParamIsMap =
      rawParams instanceof Object &&
      !(rawParams instanceof Date) &&
      !(rawParams instanceof Array);

    const params = rawParamIsMap ? rawParams : {};
    parseStatement(statement, props.lang, params, examples)
      .then((parsedStatement) => {
        return unified()
          .use(remarkParse)
          .use(remarkRehype)
          .use(rehypeStringify)
          .process(parsedStatement);
      })
      .then((newStatement) => setStatement(String(newStatement)))
      .catch((err) => console.log(err));
  }, [props]);

  return (
    <Container>
      <KatexRender text={String(statement)} />
    </Container>
  );
};

export default Statement;

export const StatementOnHttp: React.FC<{
  lang: Lang;
  baseUrl: URL;
}> = (props) => {
  const { lang, baseUrl } = props;

  const infoQuery = useQuery([baseUrl.href, "info.toml"], async () =>
    fetch(new URL("info.toml", baseUrl.href)).then((r) => {
      if (r.status == 200) {
        return r.text();
      } else {
        return Promise.reject("failed to fetch info.toml:" + r.status);
      }
    })
  );

  const statement = useQuery([baseUrl.href, "task.md"], () =>
    fetch(new URL("task.md", baseUrl.href)).then((r) => {
      if (r.status == 200) {
        return r.text();
      } else {
        return Promise.reject("failed to fetch task.md:" + r.status);
      }
    })
  );

  const info = (() => {
    if (!infoQuery.isSuccess) return {};
    try {
      return parse(infoQuery.data);
    } catch (error) {
      console.log(error);
      return {};
    }
  })();

  console.log(info);

  const exampleNumber = (() => {
    if (!info.tests) return null;
    return (info.tests as JsonMap[]).find((v) => v.name === "example.in")
      ?.number as Number;
  })();

  console.log("example", exampleNumber);

  const examples = Array.from(Array(exampleNumber), (_, k) => `example_0${k}`);
  console.log("example", examples);
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

  const data: StatementData = {
    info: info,
    statement: "",
    examples: {},
  };

  if (statement.isSuccess) {
    data.statement = statement.data;
  }

  examples.forEach((name, index) => {
    const query = inExampleQueries[index];
    if (query.isSuccess) {
      data.examples[`${name}.in`] = query.data;
    }
  });
  examples.forEach((name, index) => {
    const query = outExampleQueries[index];
    if (query.isSuccess) {
      data.examples[`${name}.out`] = query.data;
    }
  });

  console.log("fetched examples:", data.examples);

  return <Statement lang={lang} data={data} />;
};
