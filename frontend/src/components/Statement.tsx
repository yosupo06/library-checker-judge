import React, { useEffect, useState } from "react";
import { Container } from "@mui/material";
import { Lang } from "../contexts/LangContext";
import { parseStatement } from "../utils/StatementParser";
import { JsonMap } from "@iarna/toml";
import { unified } from "unified";
import remarkRehype from "remark-rehype";
import remarkParse from "remark-parse";
import rehypeStringify from "rehype-stringify";
import KatexRender from "../components/katex/KatexRender";

const Statement: React.FC<{
  lang: Lang;
  info: JsonMap;
  statement: string;
  examples: { [name: string]: string };
}> = (props) => {
  const [statement, setStatement] = useState("");

  useEffect(() => {
    const { info, statement, examples } = props;

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
