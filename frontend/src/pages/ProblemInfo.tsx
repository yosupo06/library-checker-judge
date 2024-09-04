import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import React, { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useLocalStorage } from "react-use";
import {
  useCurrentUser,
  useLangList,
  useProblemInfo,
  useSubmitMutation,
} from "../api/client_wrapper";
import { ProblemInfoResponse } from "../proto/library_checker";
import SourceEditor from "../components/SourceEditor";
import { GitHub, FlashOn, Person, Forum } from "@mui/icons-material";
import { Alert, Container } from "@mui/material";
import Statement, {
  useExamples,
  useProblemInfoTomlQuery,
  useSolveHpp,
  useStatement,
} from "../components/Statement";
import { useLang } from "../contexts/LangContext";
import { RpcError } from "@protobuf-ts/runtime-rpc";

import NotFound from "./NotFound";
import { Link as RouterLink } from "react-router-dom";
import { ProblemInfoToml } from "../utils/problem.info";
import { ProblemVersion } from "../utils/problem.storage";
import MainContainer from "../components/MainContainer";
import { LinkButton, ExternalLinkButton } from "../components/LinkButton";

const ProblemInfo: React.FC = () => {
  const { problemId } = useParams<"problemId">();
  if (!problemId) {
    return <NotFound />;
  }

  const problemInfoQuery = useProblemInfo(problemId);

  if (problemInfoQuery.isPending) {
    return (
      <Container>
        <CircularProgress />
      </Container>
    );
  }

  if (problemInfoQuery.isError) {
    return (
      <Container>
        <Alert severity="error">
          {(problemInfoQuery.error as RpcError).toString()}
        </Alert>
      </Container>
    );
  }

  return (
    <MainContainer title={problemInfoQuery.data.title}>
      <ProblemInfoBody
        problemId={problemId}
        problemInfo={problemInfoQuery.data}
      />
    </MainContainer>
  );
};
export default ProblemInfo;

const baseURL = new URL(import.meta.env.VITE_PUBLIC_BUCKET_URL);

const ProblemInfoBody: React.FC<{
  problemId: string;
  problemInfo: ProblemInfoResponse;
}> = (props) => {
  const { problemId, problemInfo } = props;

  const problemVersion = {
    name: problemId,
    version: problemInfo.version,
    testCasesVersion: problemInfo.testcasesVersion,
  };

  const infoTomlQuery = useProblemInfoTomlQuery(baseURL, problemVersion);

  if (infoTomlQuery.isPending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (infoTomlQuery.isError) {
    return (
      <Box>
        <Alert severity="error">
          {(infoTomlQuery.error as RpcError).toString()}
        </Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="body1" paragraph={true}>
        Time Limit: {problemInfo.timeLimit} sec
      </Typography>

      <UsefulLinks
        problemId={problemId}
        problemInfo={problemInfo}
        infoToml={infoTomlQuery.data}
      />
      <Divider />

      <StatementBody
        baseUrl={baseURL}
        problemVersion={problemVersion}
        infoToml={infoTomlQuery.data}
      />

      <Divider
        sx={{
          margin: 1,
        }}
      />

      <SolveHpp baseUrl={baseURL} problemVersion={problemVersion} />

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
  problemVersion: ProblemVersion;
  infoToml: ProblemInfoToml;
}> = (props) => {
  const { baseUrl, problemVersion, infoToml } = props;

  const lang = useLang();

  const statement = useStatement(baseUrl, problemVersion);

  const examples = useExamples(infoToml, baseUrl, problemVersion);

  return (
    <Statement
      lang={lang}
      data={{
        info: infoToml,
        statement: statement.isSuccess ? statement.data : "",
        examples: examples,
      }}
    />
  );
};

const UsefulLinks: React.FC<{
  problemInfo: ProblemInfoResponse;
  problemId: string;
  infoToml: ProblemInfoToml;
}> = (props) => {
  const { problemInfo, problemId, infoToml } = props;

  const currentUser = useCurrentUser();

  const fastestParams = new URLSearchParams({
    problem: problemId,
    order: "+time",
    status: "AC",
  });

  return (
    <Box>
      {currentUser.isSuccess && currentUser.data.user?.name && (
        <LinkButton
          LinkComponent={RouterLink}
          variant="outlined"
          startIcon={<Person />}
          to={`/submissions/?${new URLSearchParams({
            problem: problemId,
            user: currentUser.data.user?.name,
            status: "AC",
          }).toString()}`}
        >
          My Submissions
        </LinkButton>
      )}
      <LinkButton
        LinkComponent={RouterLink}
        variant="outlined"
        startIcon={<FlashOn />}
        to={`/submissions/?${fastestParams.toString()}`}
      >
        Fastest
      </LinkButton>
      <ExternalLinkButton startIcon={<GitHub />} href={problemInfo.sourceUrl}>
        GitHub
      </ExternalLinkButton>
      {infoToml.forum && (
        <ExternalLinkButton startIcon={<Forum />} href={infoToml.forum}>
          Forum
        </ExternalLinkButton>
      )}
    </Box>
  );
};

const SolveHpp: React.FC<{
  baseUrl: URL;
  problemVersion: ProblemVersion;
}> = (props) => {
  const { baseUrl, problemVersion } = props;

  const solveHppQuery = useSolveHpp(baseUrl, problemVersion);

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

  if (langListQuery.isPending) {
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
        <FormControl fullWidth>
          <SourceEditor
            value={source}
            language={progLang}
            onChange={(e) => {
              setSource(e);
            }}
            readOnly={false}
            height={400}
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
