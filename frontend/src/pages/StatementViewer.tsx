import Typography from "@mui/material/Typography";
import React, { useEffect, useState } from "react";
import {
  Box,
  Button,
  Container,
  Divider,
  FormControl,
  FormHelperText,
  Stack,
  Tab,
  Tabs,
} from "@mui/material";
import { useLang } from "../contexts/LangContext";
import { parseStatement } from "../utils/StatementParser";
import { parse } from "toml";
import { unified } from "unified";
import remarkRehype from "remark-rehype";
import remarkParse from "remark-parse";
import rehypeStringify from "rehype-stringify";
import KatexTypography from "../components/katex/KatexTypography";
import SourceEditor from "../components/SourceEditor";
import KatexRender from "../components/katex/KatexRender";

const StatementSideView: React.FC<{
  enHtml: string;
  jaHtml: string;
  title: string;
}> = (props) => {
  const { enHtml, jaHtml, title } = props;
  const [tabIndex, setTabIndex] = React.useState(0);

  return (
    <Container>
      <KatexTypography variant="h2" paragraph={true}>
        {title}
      </KatexTypography>

      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={tabIndex}
          onChange={(event, newValue) => setTabIndex(newValue)}
          aria-label="basic tabs example"
        >
          <Tab label="Statement(en)" />
          <Tab label="Statement(ja)" />
        </Tabs>
      </Box>

      {tabIndex === 0 && (
        <Box sx={{ p: 3 }}>
          <KatexRender text={enHtml} />
        </Box>
      )}
      {tabIndex === 1 && (
        <Box sx={{ p: 3 }}>
          <KatexRender text={jaHtml} />
        </Box>
      )}
    </Container>
  );
};

const StatementViewer: React.FC = () => {
  const lang = useLang();

  const [tabIndex, setTabIndex] = React.useState(0);

  const [files, setFiles] = useState<File[]>([]);

  const [taskMd, setTaskMd] = useState(""); // task.md
  const [infoToml, setInfoToml] = useState(""); // info.toml
  const [infoValue, setInfoValue] = useState<unknown>({}); // parsed info.toml

  const [examples, setExamples] = useState<{ [key: string]: string }>({}); // example.in / example.out

  const [parsedEnMarkdown, setParsedEnMarkdown] = useState("");
  const [parsedJaMarkdown, setParsedJaMarkdown] = useState("");
  const [parsedEnHtml, setParsedEnHtml] = useState("");
  const [parsedJaHtml, setParsedJaHtml] = useState("");

  useEffect(() => {
    const taskFile = files.find((e) =>
      e.webkitRelativePath.endsWith("task.md")
    );
    if (!taskFile) {
      console.log("task.md not found");
      return;
    }
    taskFile.text().then((text) => {
      setTaskMd(text);
    });
  }, [files]);

  useEffect(() => {
    const infoFile = files.find((e) =>
      e.webkitRelativePath.endsWith("info.toml")
    );
    if (!infoFile) {
      console.log("info.toml not found");
      return;
    }
    infoFile.text().then((text) => {
      setInfoToml(text);
    });
  }, [files]);

  useEffect(() => {
    console.log(parse(infoToml));
    setInfoValue(parse(infoToml));
  }, [infoToml]);

  useEffect(() => {
    const pattern = /(in|out)\/example_[0-9]*.(in|out)/;
    const exampleFiles = files.filter((e) =>
      e.webkitRelativePath.match(pattern)
    );
    Promise.all(exampleFiles.map((e) => e.text())).then((texts) => {
      const examples: { [key: string]: string } = {};
      exampleFiles.forEach((value, index) => {
        examples[value.name] = texts[index];
      });
      setExamples(examples);
    });
  }, [files]);

  useEffect(() => {
    parseStatement(taskMd, "en", infoValue.params, examples).then(
      setParsedEnMarkdown
    );
    parseStatement(taskMd, "ja", infoValue.params, examples).then(
      setParsedJaMarkdown
    );
  }, [lang, taskMd, infoValue, examples]);

  useEffect(() => {
    unified()
      .use(remarkParse)
      .use(remarkRehype)
      .use(rehypeStringify)
      .process(parsedEnMarkdown)
      .then((e) => setParsedEnHtml(String(e)));
    unified()
      .use(remarkParse)
      .use(remarkRehype)
      .use(rehypeStringify)
      .process(parsedJaMarkdown)
      .then((e) => setParsedJaHtml(String(e)));
  }, [parsedEnMarkdown, parsedJaMarkdown]);

  const editorSideView = (
    <Container>
      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={tabIndex}
          onChange={(event, newValue) => setTabIndex(newValue)}
          aria-label="basic tabs example"
        >
          <Tab label="task.md" />
          <Tab label="info.toml" />
        </Tabs>
      </Box>

      {tabIndex === 0 && (
        <Box sx={{ p: 3 }}>
          <Typography variant="h4" paragraph={true}>
            task.md
          </Typography>
          <Box
            sx={{
              height: "400px",
              width: "100%",
            }}
          >
            <SourceEditor
              value={taskMd}
              language="plaintext" // todo: markdown?
              onChange={(e) => {
                setTaskMd(e);
              }}
              readOnly={false}
              autoHeight={false}
            />
          </Box>
        </Box>
      )}
      {tabIndex === 1 && (
        <Box sx={{ p: 3 }}>
          <Typography variant="h4" paragraph={true}>
            info.toml
          </Typography>
          <Box
            sx={{
              height: "400px",
              width: "100%",
            }}
          >
            <SourceEditor
              value={infoToml}
              language="plaintext" // todo: markdown?
              onChange={(e) => {
                setInfoToml(e);
              }}
              readOnly={false}
              autoHeight={false}
            />
          </Box>
        </Box>
      )}
    </Container>
  );

  return (
    <Box>
      <Container maxWidth="xl">
        <Typography variant="h2" paragraph={true}>
          Statement Viewer
        </Typography>

        <FormControl>
          <Button
            variant="outlined"
            component="label"
            sx={{
              width: 150,
            }}
          >
            Load directory
            <input
              hidden
              type="file"
              onChange={(e) => {
                setFiles(Array.from(e.target.files || []));
              }}
              /* @ts-expect-error webkitdirectory is not standard */
              webkitdirectory=""
            />
          </Button>
          <FormHelperText>
            set problem directory (example:
            /path/to/library-checker-problems/sample/aplusb)
          </FormHelperText>
        </FormControl>
      </Container>

      <Box>
        <Divider
          sx={{
            margin: 3,
          }}
        />
        <Stack
          direction="row"
          divider={<Divider orientation="vertical" flexItem />}
          spacing={2}
        >
          {editorSideView}
          <StatementSideView
            enHtml={parsedEnHtml}
            jaHtml={parsedJaHtml}
            title={infoValue["title"]}
          />
        </Stack>
      </Box>
    </Box>
  );
};

export default StatementViewer;
