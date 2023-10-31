import { Liquid } from "liquidjs";
import { TagToken, Template, ParseStream } from "liquidjs";
import { Lang } from "../contexts/LangContext";
import { AnyJson } from "@iarna/toml";

export const parseStatement = (
  statement: string,
  lang: Lang,
  params: { [key in string]: AnyJson },
  examples: { [name in string]: string }
): Promise<string> => {
  return engine.parseAndRender(statement, {
    lang: lang,
    params: params,
    examples: examples,
  });
};

const engine = new Liquid({
  tagDelimiterLeft: "@{",
  tagDelimiterRight: "}",
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

const numberParamToStr = (value: number) => {
  if (Number.isInteger(value)) {
    if (value != 0 && value % 1000000 == 0) {
      const k = Math.floor(Math.log10(Math.abs(value)));
      if (value === 10 ** k) {
        return `10^{${k}}`;
      }
    }
  }
  return value.toString();
};

const paramToStr = (value: unknown) => {
  if (typeof value === "number") {
    return numberParamToStr(value);
  } else {
    return String(value);
  }
};

engine.registerTag("keyword", {
  parse(tagToken) {
    this.value = tagToken.args.substring(1);
  },
  render(context, emitter) {
    const targetLang = context.get(["lang"]) as unknown as Lang;
    if (this.value in keywordsDict) {
      emitter.write(keywordsDict[this.value][targetLang]);
    }
  },
});

engine.registerTag("param", {
  parse(tagToken) {
    this.value = tagToken.args.substring(1);
  },
  render(context, emitter) {
    const params = context.get(["params"]) as { [key: string]: object };
    emitter.write(paramToStr(params[this.value]));
  },
});

engine.registerTag("example", {
  parse(tagToken) {
    this.value = tagToken.args.substring(1);
    this.exampleCounter = 1;
  },
  render(context, emitter) {
    const examples = context.get(["examples"]) as { [key: string]: object };
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
    const targetLang = context.get(["lang"]);
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
