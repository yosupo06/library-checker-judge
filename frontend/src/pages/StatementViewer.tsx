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
  TextField,
} from "@mui/material";
import KatexTypography from "../components/katex/KatexTypography";
import SourceEditor from "../components/SourceEditor";
import Statement from "../components/Statement";
import { parse } from "@iarna/toml";
import urlJoin from "url-join";
import { SubmitHandler, useForm } from "react-hook-form";
import { ProblemInfoToml, parseProblemInfoToml } from "../utils/problem.info";

type StatementData = {
  info: string;
  statement: string;
  examples: { [name: string]: string };
};

const StatementViewer: React.FC = () => {
  const [data, setData] = useState<StatementData | null>({
    info: "",
    statement: "",
    examples: {},
  });

  return (
    <Box>
      <Container maxWidth="xl">
        <Typography variant="h2" paragraph={true}>
          Statement Viewer
        </Typography>

        <DataLoader updateData={(newData) => setData(newData)} />
      </Container>

      <Box>
        <Divider
          sx={{
            margin: 3,
          }}
        />

        {data && <StatementViewerInternal data={data} />}
      </Box>
    </Box>
  );
};

export default StatementViewer;

const DataLoader: React.FC<{
  updateData: (data: StatementData) => void;
}> = (props) => {
  const [data, setData] = useState<StatementData>({
    info: "",
    statement: "",
    examples: {},
  });

  useEffect(() => {
    props.updateData(data);
  }, [data]);

  const setFiles = (files: File[]) => {
    const taskFile = files.find((e) =>
      e.webkitRelativePath.endsWith("task.md")
    );
    if (!taskFile) {
      console.log("task.md not found");
      return;
    }
    taskFile
      .text()
      .then((text) => {
        setData((data) => ({
          ...data,
          statement: text,
        }));
      })
      .catch((err) => {
        console.log(err);
      });

    const infoFile = files.find((e) =>
      e.webkitRelativePath.endsWith("info.toml")
    );
    if (!infoFile) {
      console.log("info.toml not found");
      return;
    }
    infoFile.text().then((text) => {
      setData((data) => ({
        ...data,
        info: text,
      }));
    });

    const pattern = /(in|out)\/example_[0-9]*.(in|out)/;
    const exampleFiles = files.filter((e) =>
      e.webkitRelativePath.match(pattern)
    );
    Promise.all(exampleFiles.map((e) => e.text())).then((texts) => {
      const examples: { [key: string]: string } = {};
      exampleFiles.forEach((value, index) => {
        examples[value.name] = texts[index];
      });
      setData((data) => ({
        ...data,
        examples: examples,
      }));
    });
  };

  return (
    <>
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
      <Divider sx={{ margin: 3 }} />
      <GithubDataLoader updateData={setData} />
    </>
  );
};

const GithubDataLoader: React.FC<{
  updateData: (data: StatementData) => void;
}> = (props) => {
  const [data, setData] = useState<StatementData>({
    info: "",
    statement: "",
    examples: {},
  });

  type Form = {
    ref: string;
    problem: string;
  };
  const { register, handleSubmit } = useForm<Form>();

  useEffect(() => {
    props.updateData(data);
  }, [data]);

  const onSubmit: SubmitHandler<Form> = (data) => {
    const { ref, problem } = data;
    const [owner, branch] = ref.split(":");
    const baseUrl = urlJoin(
      "https://raw.githubusercontent.com/",
      owner,
      "library-checker-problems",
      branch,
      problem
    );

    fetch(new URL(urlJoin(baseUrl, "info.toml")))
      .then((r) => {
        if (r.status == 200) {
          return r.text();
        } else {
          throw new Error("failed to fetch info.toml:" + r.status);
        }
      })
      .then((info) => {
        setData((data) => ({
          ...data,
          info: info,
        }));
      });

    fetch(new URL(urlJoin(baseUrl, "task.md")))
      .then((r) => {
        if (r.status == 200) {
          return r.text();
        } else {
          throw new Error("failed to fetch task.md:" + r.status);
        }
      })
      .then((task) => {
        setData((data) => ({
          ...data,
          statement: task,
        }));
      });
  };

  return (
    <>
      <FormControl>
        <Stack
          direction="row"
          divider={<Divider orientation="vertical" flexItem />}
          spacing={2}
        >
          <TextField label="yosupo06:master" {...register("ref")} />
          <TextField label="sample/aplusb" {...register("problem")} />
          <Box
            sx={{
              display: "flex",
            }}
          >
            <Button
              variant="outlined"
              component="label"
              sx={{
                width: 150,
                marginTop: "auto",
              }}
              onClick={handleSubmit(onSubmit)}
            >
              Load from GitHub
            </Button>
          </Box>
        </Stack>
      </FormControl>
    </>
  );
};

interface Props {
  data: StatementData;
}

const StatementViewerInternal: React.FC<Props> = (props) => {
  const [editorTabIndex, setEditorTabIndex] = useState(0);
  const [viewerTabIndex, setViewerTabIndex] = useState(0);
  const [info, setInfo] = useState(props.data.info);
  const [statement, setStatement] = useState(props.data.statement);

  useEffect(() => {
    setInfo(props.data.info);
  }, [props.data.info]);
  useEffect(() => {
    setStatement(props.data.statement);
  }, [props.data.statement]);

  const infoToml = (() => {
    try {
      return parseProblemInfoToml(info);
    } catch (error) {
      console.log(error);
      return {
        tests: [],
        params: {},
      };
    }
  })();

  const editorSide = (
    <Container>
      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={editorTabIndex}
          onChange={(event, newValue) => setEditorTabIndex(newValue)}
          aria-label="basic tabs example"
        >
          <Tab label="task.md" />
          <Tab label="info.toml" />
        </Tabs>
      </Box>

      {editorTabIndex === 0 && (
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
              value={statement}
              language="plaintext" // todo: markdown?
              onChange={(e) => {
                setStatement(e);
              }}
              readOnly={false}
              autoHeight={false}
            />
          </Box>
        </Box>
      )}
      {editorTabIndex === 1 && (
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
              value={info}
              language="plaintext"
              onChange={(e) => {
                setInfo(e);
              }}
              readOnly={false}
              autoHeight={false}
            />
          </Box>
        </Box>
      )}
    </Container>
  );

  const viewerSide = (
    <Container>
      <KatexTypography variant="h2" paragraph={true}>
        {infoToml?.title ?? "<title not found>"}
      </KatexTypography>

      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={viewerTabIndex}
          onChange={(_, newValue) => setViewerTabIndex(newValue)}
          aria-label="basic tabs example"
        >
          <Tab label="Statement(en)" />
          <Tab label="Statement(ja)" />
        </Tabs>
      </Box>

      {viewerTabIndex === 0 && (
        <Box sx={{ p: 3 }}>
          <Statement
            lang="en"
            data={{
              info: infoToml,
              statement: statement,
              examples: props.data.examples,
            }}
          />
        </Box>
      )}
      {viewerTabIndex === 1 && (
        <Box sx={{ p: 3 }}>
          <Statement
            lang="ja"
            data={{
              info: infoToml,
              statement: statement,
              examples: props.data.examples,
            }}
          />
        </Box>
      )}
    </Container>
  );

  return (
    <Stack
      direction="row"
      divider={<Divider orientation="vertical" flexItem />}
      spacing={2}
    >
      {editorSide}
      {viewerSide}
    </Stack>
  );
};
