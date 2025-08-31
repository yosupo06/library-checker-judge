import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import FormControlLabel from "@mui/material/FormControlLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Switch from "@mui/material/Switch";
import Stack from "@mui/material/Stack";
import InputLabel from "@mui/material/InputLabel";
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
import { useTranslation } from "../utils/translations";
import { Alert, Container } from "@mui/material";
import Statement from "../components/Statement";
import { useLang } from "../contexts/LangContext";
import { RpcError } from "@protobuf-ts/runtime-rpc";

import NotFound from "./NotFound";
import { Link as RouterLink } from "react-router-dom";
import { ProblemInfoToml } from "../utils/problem.info";
import { useProblemAssets } from "../utils/problem.storage";
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

  const assets = useProblemAssets(baseURL, problemId, problemInfo);

  if (assets.isPending || !assets.info) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (assets.error) {
    return (
      <Box>
        <Alert severity="error">
          {String(assets.error)}
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
        infoToml={assets.info}
      />
      <Divider />

      <StatementBody info={assets.info} statement={assets.statement ?? ""} examples={assets.examples} />

      <Divider
        sx={{
          margin: 1,
        }}
      />

      <SolveHpp solveHpp={assets.solveHpp ?? null} />

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
  info: ProblemInfoToml;
  statement: string;
  examples: { [name: string]: string };
}> = (props) => {
  const { info, statement, examples } = props;
  const lang = useLang();
  return <Statement lang={lang} data={{ info, statement, examples }} />;
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

const SolveHpp: React.FC<{ solveHpp: string | null }> = (props) => {
  const { solveHpp } = props;
  return (
    <Box>
      <Typography variant="h4" paragraph={true}>
        C++(Function) header
      </Typography>
      {solveHpp && (
        <>
          <Typography variant="h6" paragraph={true}>
            solve.hpp
          </Typography>
          <SourceEditor value={solveHpp} language="cpp" readOnly={true} />
        </>
      )}
      {solveHpp === null && (
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
  const [tleKnockout, setTleKnockout] = useState(true);

  const lang = useLang();
  const t = useTranslation(lang);

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
      tleKnockout: tleKnockout,
    });
  };

  return (
    <Box>
      <Typography variant="h4" paragraph={true}>
        Submit
      </Typography>

      <form onSubmit={handleSubmit}>
        <Stack spacing={3}>
          {/* Language Selection */}
          <FormControl fullWidth>
            <InputLabel id="language-select-label">
              {t("languageLabel")}
            </InputLabel>
            <Select
              labelId="language-select-label"
              displayEmpty
              required
              value={progLang}
              onChange={(e) => setProgLang(e.target.value as string)}
              label={t("languageLabel")}
            >
              <MenuItem value="">{t("language")}</MenuItem>
              {langListQuery.data.langs.map((e) => (
                <MenuItem key={e.id} value={e.id}>
                  {e.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          {/* TLE Knockout Option */}
          <FormControl>
            <FormControlLabel
              control={
                <Switch
                  checked={tleKnockout}
                  onChange={(e) => setTleKnockout(e.target.checked)}
                  color="primary"
                />
              }
              label={t("tleKnockoutLabel")}
            />
          </FormControl>

          {/* Source Code Editor */}
          <FormControl fullWidth>
            <SourceEditor
              value={source}
              language={progLang}
              onChange={(e) => {
                setSource(e);
              }}
              readOnly={false}
              height={400}
              placeholder="Enter your solution code here..."
            />
          </FormControl>

          {/* Submit Button */}
          <Box sx={{ display: "flex", justifyContent: "center", mt: 2 }}>
            <Button
              variant="outlined"
              color="primary"
              type="submit"
              size="large"
              sx={{
                px: 4,
                py: 1.5,
                fontWeight: "bold",
                borderWidth: 2,
                "&:hover": {
                  borderWidth: 2,
                },
              }}
            >
              {t("submit")}
            </Button>
          </Box>
        </Stack>
      </form>
    </Box>
  );
};
