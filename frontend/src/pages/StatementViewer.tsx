import Typography from "@mui/material/Typography";
import React, { useState } from "react";
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
import urlJoin from "url-join";
import { SubmitHandler, useForm } from "react-hook-form";
import { parseProblemInfoToml } from "../utils/problem.info";
import { StatementData } from "../utils/statement.parser";

type RawStatementData = {
  info: string;
  statement: string;
  examples: { [name: string]: string };
};

const StatementViewer: React.FC = () => {
  const [data, setData] = useState<RawStatementData>({
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

        <DataLoader setData={setData} />
      </Container>

      <Box>
        <Divider
          sx={{
            margin: 3,
          }}
        />

        <StatementViewerInternal data={data} setData={setData} />
      </Box>
    </Box>
  );
};

export default StatementViewer;

const DataLoader: React.FC<{
  setData: (f: (data: RawStatementData) => RawStatementData) => void;
}> = (props) => {
  const { setData } = props;
  return (
    <>
      <FileLoader setData={setData} />
      <Divider sx={{ margin: 3 }} />
      <GithubDataLoader setData={setData} />
    </>
  );
};

const FileLoader: React.FC<{
  setData: (f: (data: RawStatementData) => RawStatementData) => void;
}> = (props) => {
  const { setData } = props;

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
    </>
  );
};

const GithubDataLoader: React.FC<{
  setData: (f: (data: RawStatementData) => RawStatementData) => void;
}> = (props) => {
  const { setData } = props;

  type Form = {
    ref: string;
    problem: string;
  };
  const { register, handleSubmit } = useForm<Form>();

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

const StatementViewerInternal: React.FC<{
  data: RawStatementData;
  setData: (d: RawStatementData) => void;
}> = (props) => {
  const { data, setData } = props;

  return (
    <Stack
      direction="row"
      divider={<Divider orientation="vertical" flexItem />}
      spacing={2}
    >
      <EditorSide data={data} setData={setData} />
      <ViewerSide
        data={{
          info: parseProblemInfoToml(data.info),
          statement: data.statement,
          examples: data.examples,
        }}
      />
    </Stack>
  );
};

const EditorSide: React.FC<{
  data: RawStatementData;
  setData: (d: RawStatementData) => void;
}> = (props) => {
  const { data, setData } = props;

  const setInfo = (info: string) => {
    setData({
      ...data,
      info: info,
    });
  };
  const setStatement = (statement: string) => {
    setData({
      ...data,
      statement: statement,
    });
  };

  const [tabIndex, setTabIndex] = useState(0);

  return (
    <Container>
      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={tabIndex}
          onChange={(_, newValue) => setTabIndex(newValue)}
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
              value={data.statement}
              language="markdown"
              onChange={setStatement}
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
              value={data.info}
              language="plaintext"
              onChange={setInfo}
              readOnly={false}
              autoHeight={false}
            />
          </Box>
        </Box>
      )}
    </Container>
  );
};

const ViewerSide: React.FC<{
  data: StatementData;
}> = (props) => {
  const { data } = props;
  const [tabIndex, setTabIndex] = useState(0);

  return (
    <Container>
      <KatexTypography variant="h2" paragraph={true}>
        {data.info.title ?? "<title not found>"}
      </KatexTypography>

      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={tabIndex}
          onChange={(_, newValue) => setTabIndex(newValue)}
        >
          <Tab label="Statement(en)" />
          <Tab label="Statement(ja)" />
        </Tabs>
      </Box>

      {tabIndex === 0 && (
        <Box sx={{ p: 3 }}>
          <Statement lang="en" data={data} />
        </Box>
      )}
      {tabIndex === 1 && (
        <Box sx={{ p: 3 }}>
          <Statement lang="ja" data={data} />
        </Box>
      )}
    </Container>
  );
};
