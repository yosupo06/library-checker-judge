import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import React, { useState } from "react";
import { LinkProps, useNavigate, useParams } from "react-router-dom";
import { useLocalStorage } from "react-use";
import {
  useCurrentUser,
  useLangList,
  useProblemInfo,
  useSubmitMutation,
} from "../api/client_wrapper";
import { ProblemInfoResponse } from "../proto/library_checker";
import SourceEditor from "../components/SourceEditor";
import { GitHub, FlashOn, Person } from "@mui/icons-material";
import KatexTypography from "../components/katex/KatexTypography";
import { Alert, Container } from "@mui/material";
import Statement, {
  useExamples,
  useProblemInfoToml,
  useSolveHpp,
  useStatement,
} from "../components/Statement";
import { useLang } from "../contexts/LangContext";
import urlJoin from "url-join";
import { RpcError } from "@protobuf-ts/runtime-rpc";

import NotFound from "./NotFound";
import { Link as RouterLink } from "react-router-dom";
import styled from "@emotion/styled";
import { ProblemInfoToml } from "../utils/problem.info";

const ProblemInfo: React.FC = () => {
  const { problemId } = useParams<"problemId">();
  if (!problemId) {
    return <NotFound />;
  }

  return (
    <Container>
      <ProblemInfoBody problemId={problemId} />
    </Container>
  );
};
export default ProblemInfo;

const ProblemInfoBody: React.FC<{ problemId: string }> = (props) => {
  const { problemId } = props;

  const problemInfoQuery = useProblemInfo(problemId);

  if (problemInfoQuery.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (problemInfoQuery.isError) {
    return (
      <Box>
        <Alert severity="error">
          {(problemInfoQuery.error as RpcError).toString()}
        </Alert>
      </Box>
    );
  }

  const version = problemInfoQuery.data.version;

  const baseUrl = new URL(
    urlJoin(import.meta.env.VITE_PUBLIC_BUCKET_URL, `${problemId}/${version}/`)
  );

  return (
    <Box>
      <KatexTypography variant="h2" paragraph={true}>
        {problemInfoQuery.data.title}
      </KatexTypography>

      <Typography variant="body1" paragraph={true}>
        Time Limit: {problemInfoQuery.data.timeLimit} sec
      </Typography>

      <UsefulLinks problemId={problemId} problemInfo={problemInfoQuery.data} />
      <Divider />

      <StatementBody baseUrl={baseUrl} />

      <Divider
        sx={{
          margin: 1,
        }}
      />

      <SolveHpp baseUrl={baseUrl} />

      <Divider
        sx={{
          margin: 1,
        }}
      />

      <SubmitForm problemId={problemId} />
    </Box>
  );
};

export const StatementBody: React.FC<{
  baseUrl: URL;
}> = (props) => {
  const { baseUrl } = props;

  const lang = useLang();

  const statement = useStatement(baseUrl);
  const infoToml = useProblemInfoToml(baseUrl);
  const info: ProblemInfoToml = infoToml.isSuccess
    ? infoToml.data
    : {
        tests: [],
        params: {},
      };

  const examples = useExamples(info, baseUrl);

  return (
    <Statement
      lang={lang}
      data={{
        info: info,
        statement: statement.isSuccess ? statement.data : "",
        examples: examples,
      }}
    />
  );
};

const UsefulLinks: React.FC<{
  problemInfo: ProblemInfoResponse;
  problemId: string;
}> = (props) => {
  const ButtonLink = styled(Button)<LinkProps>();

  const { problemInfo, problemId } = props;

  const currentUser = useCurrentUser();

  const fastestParams = new URLSearchParams({
    problem: problemId,
    order: "+time",
    status: "AC",
  });

  return (
    <Box>
      {currentUser.isSuccess && currentUser.data.user?.name && (
        <ButtonLink
          LinkComponent={RouterLink}
          variant="outlined"
          startIcon={<Person />}
          to={`/submissions/?${new URLSearchParams({
            problem: problemId,
            user: currentUser.data.user?.name,
            status: "AC",
          }).toString()}`}
        >
          My Submission
        </ButtonLink>
      )}
      <ButtonLink
        LinkComponent={RouterLink}
        variant="outlined"
        startIcon={<FlashOn />}
        to={`/submissions/?${fastestParams.toString()}`}
      >
        Fastest
      </ButtonLink>
      <Button
        variant="outlined"
        startIcon={<GitHub />}
        href={problemInfo.sourceUrl}
      >
        Github
      </Button>
    </Box>
  );
};

const SolveHpp: React.FC<{ baseUrl: URL }> = (props) => {
  const { baseUrl } = props;

  const solveHppQuery = useSolveHpp(baseUrl);

  return (
    <Box>
      <Typography variant="h4" paragraph={true}>
        C++(Function) header
      </Typography>

      {solveHppQuery.isSuccess && solveHppQuery.data && (
        <>
          <Typography variant="h6" paragraph={true}>
            solve.hpp
          </Typography>
          <SourceEditor
            value={solveHppQuery.data}
            language="cpp"
            readOnly={true}
            autoHeight={true}
          />
        </>
      )}
      {solveHppQuery.isSuccess && !solveHppQuery.data && (
        <Typography variant="body1" paragraph={true}>
          Unsupported
        </Typography>
      )}
    </Box>
  );
};

const SubmitForm: React.FC<{ problemId: string }> = (props) => {
  const { problemId } = props;
  const navigate = useNavigate();
  const [source, setSource] = useState("");
  const [progLang, setProgLang] = useLocalStorage("programming-lang", "");

  const langListQuery = useLangList();

  const submitMutation = useSubmitMutation({
    onSuccess: (resp) => {
      navigate(`/submission/${resp.id}`);
    },
  });

  if (langListQuery.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (langListQuery.isError) {
    return (
      <Box>
        <Alert severity="error">
          {(langListQuery.error as RpcError).toString()}
        </Alert>
      </Box>
    );
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!progLang) {
      throw new Error("programming language must be set");
    }

    submitMutation.mutate({
      lang: progLang,
      problem: problemId,
      source: source,
    });
  };

  return (
    <Box>
      <Typography variant="h4" paragraph={true}>
        Submit
      </Typography>

      <form onSubmit={handleSubmit}>
        <FormControl
          sx={{
            height: "400px",
            width: "100%",
          }}
        >
          <SourceEditor
            value={source}
            language={progLang}
            onChange={(e) => {
              setSource(e);
            }}
            readOnly={false}
            autoHeight={false}
          />
        </FormControl>
        <FormControl>
          <Select
            displayEmpty
            required
            value={progLang}
            onChange={(e) => setProgLang(e.target.value as string)}
          >
            <MenuItem value="">Lang</MenuItem>
            {langListQuery.data.langs.map((e) => (
              <MenuItem key={e.id} value={e.id}>
                {e.name}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <Button color="primary" type="submit">
          Submit
        </Button>
      </form>
    </Box>
  );
};
