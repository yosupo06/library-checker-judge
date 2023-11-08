import { Context, Liquid } from "liquidjs";
import { TagToken, Template, ParseStream } from "liquidjs";
import { Lang } from "../contexts/LangContext";
import { ProblemInfoToml } from "./problem.info";
import { useQuery } from "@tanstack/react-query";
import { unified } from "unified";
import remarkRehype from "remark-rehype";
import remarkParse from "remark-parse";
import rehypeStringify from "rehype-stringify";

export type StatementData = {
  info: ProblemInfoToml;
  statement: string;
  examples: { [name: string]: string };
};

export const useStatementParser = (lang: Lang, data: StatementData) => {
  return useQuery({
    queryKey: ["statement", "parser", lang, data],
    queryKeyHashFn: (key) =>
      JSON.stringify(key, (_, v) => (typeof v === "bigint" ? v.toString() : v)),
    queryFn: () =>
      parseCustomTag(lang, data)
        .then(parseMarkdown)
        .then((statement) => String(statement)),
  });
};

const parseMarkdown = (mdStatement: string) =>
  unified()
    .use(remarkParse)
    .use(remarkRehype)
    .use(rehypeStringify)
    .process(mdStatement);

const parseCustomTag = (lang: Lang, data: StatementData): Promise<string> => {
  return engine.parseAndRender(data.statement, {
    lang: lang,
    data: data,
    params: data.info.params,
    examples: data.examples,
  });
};

const engine = new Liquid({
  tagDelimiterLeft: "@{",
  tagDelimiterRight: "}",
  outputDelimiterLeft: "!WEDONTUSETHISFUNCTION!",
  outputDelimiterRight: "!WEDONTUSETHISFUNCTION!",
});

const keywordsDict: { [key: string]: { [lang in Lang]: string } } = {
  statement: {
    en: "Problem Statement",
    ja: "問題文",
  },
  constraints: {
    en: "Constraints",
    ja: "制約",
  },
  input: {
    en: "Input",
    ja: "入力",
  },
  output: {
    en: "Output",
    ja: "出力",
  },
  sample: {
    en: "Sample",
    ja: "サンプル",
  },
};

export const paramToStr = (value: bigint) => {
  if (value == 0n) return "0";

  if (value % 100_000n == 0n) {
    let rem_value = value;
    let k = 0;
    while (rem_value % 10n == 0n) {
      rem_value /= 10n;
      k++;
    }

    if (rem_value == 1n) {
      return `10^{${k}}`;
    } else {
      return `${rem_value} \\times 10^{${k}}`;
    }
  }

  return value.toString();
};

const getLang = (context: Context) => {
  const lang = context.getSync(["lang"]);

  if (typeof lang === "string") {
    if (lang == "en") return "en";
    if (lang == "ja") return "ja";
  }
  return "en";
};
const getStatementData = (context: Context) =>
  context.getSync(["data"]) as StatementData;

engine.registerTag("keyword", {
  parse(tagToken) {
    this.value = tagToken.args.substring(1);
  },
  render(context, emitter) {
    const lang = getLang(context);
    if (this.value in keywordsDict) {
      emitter.write(keywordsDict[this.value][lang]);
    }
  },
});

engine.registerTag("param", {
  parse(tagToken) {
    this.value = tagToken.args.substring(1);
  },
  render(context, emitter) {
    const params = getStatementData(context).info.params;
    emitter.write(paramToStr(params[this.value]));
  },
});

engine.registerTag("example", {
  parse(tagToken) {
    this.value = tagToken.args.substring(1);
    this.exampleCounter = 1;
  },
  render(context, emitter) {
    const examples = getStatementData(context).examples;

    const inName = this.value + ".in";
    const outName = this.value + ".out";

    emitter.write(`### #${this.exampleCounter}\n`);
    this.exampleCounter++;

    emitter.write("```\n");
    if (inName in examples) {
      emitter.write(examples[inName]);
    } else {
      emitter.write(`${inName} not found!\n`);
    }
    emitter.write("```\n");

    emitter.write("```\n");
    if (outName in examples) {
      emitter.write(examples[outName]);
    } else {
      emitter.write(`${outName} not found!\n`);
    }
    emitter.write("```\n");
  },
});

engine.registerTag("lang", {
  parse(tagToken, remainTokens) {
    this.sections = [];
    let currentLang = tagToken.args.substring(1);
    let p: Template[] = [];
    const stream: ParseStream = this.liquid.parser
      .parseStream(remainTokens)
      .on("tag:lang", (token: TagToken) => {
        this.sections.push({
          lang: currentLang,
          templates: p,
        });
        p = [];
        currentLang = token.args.substring(1);
        if (currentLang === "end") {
          stream.stop();
          return;
        }
      })
      .on("template", (tpl: Template) => p.push(tpl))
      .on("end", () => {
        throw new Error(`tag ${tagToken.getText()} not closed`);
      });

    stream.start();
  },
  *render(context, emitter) {
    const targetLang = getLang(context);
    for (const section of this.sections) {
      if (section.lang === targetLang) {
        yield this.liquid.renderer.renderTemplates(
          section.templates,
          context,
          emitter
        );
      }
    }
  },
});
