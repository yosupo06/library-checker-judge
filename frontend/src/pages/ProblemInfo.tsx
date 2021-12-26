import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import React, { useContext, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useLocalStorage } from "react-use";
import library_checker_client, {
  authMetadata,
  useLangList,
  useProblemInfo,
} from "../api/library_checker_client";
import { ProblemInfoResponse, SubmitRequest } from "../api/library_checker_pb";
import SourceEditor from "../components/SourceEditor";
import KatexRender from "../components/katex/KatexRender";
import { AuthContext } from "../contexts/AuthContext";
import { GitHub, FlashOn, Person } from "@mui/icons-material";
import { styled } from "@mui/system";
import KatexTypography from "../components/katex/KatexTypography";

const PlainLink = styled(Link)({
  color: "inherit",
  textDecoration: "none",
  textTransform: "none",
});

const UsefulLinks: React.FC<{
  problemInfo: ProblemInfoResponse;
  problemId: string;
  userId: string | undefined;
}> = (props) => {
  const { problemInfo, problemId, userId } = props;
  const fastestParams = new URLSearchParams({
    problem: problemId,
    order: "+time",
    status: "AC",
  });

  return (
    <Box>
      {userId && (
        <Button variant="outlined" startIcon={<Person />}>
          <PlainLink
            to={`/submissions/?${new URLSearchParams({
              problem: problemId,
              user: userId,
              status: "AC",
            }).toString()}`}
          >
            My Submissions
          </PlainLink>
        </Button>
      )}
      <Button variant="outlined" startIcon={<FlashOn />}>
        <PlainLink to={`/submissions/?${fastestParams.toString()}`}>
          Fastest
        </PlainLink>
      </Button>
      <Button
        variant="outlined"
        startIcon={<GitHub />}
        href={problemInfo.getSourceUrl()}
      >
        Github
      </Button>
    </Box>
  );
};

const ProblemInfo: React.FC = () => {
  const navigate = useNavigate();
  const auth = useContext(AuthContext);
  const [source, setSource] = useState("");
  const [lang, setLang] = useLocalStorage("programming-lang", "");

  const { problemId } = useParams<"problemId">();
  if (problemId === undefined) {
    throw new Error(`problemId is not defined`);
  }
  const problemInfoQuery = useProblemInfo(problemId);
  const langListQuery = useLangList();

  if (problemInfoQuery.isLoading || problemInfoQuery.isIdle) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (problemInfoQuery.isError || langListQuery.isError) {
    return (
      <Box>
        <Typography variant="body1">
          Error : {problemInfoQuery.error} {langListQuery.error}
        </Typography>
      </Box>
    );
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!lang) {
      console.log("Please select lang");
      return;
    }
    library_checker_client
      .submit(
        new SubmitRequest()
          .setLang(lang)
          .setProblem(problemId)
          .setSource(source),
        (auth && authMetadata(auth.state)) ?? null
      )
      .then((resp) => {
        navigate(`/submission/${resp.getId()}`);
      });
  };

  return (
    <Box>
      <KatexTypography variant="h2" paragraph={true}>
        {problemInfoQuery.data.getTitle()}
      </KatexTypography>
      <Typography variant="body1" paragraph={true}>
        Time Limit: {problemInfoQuery.data.getTimeLimit()} sec
      </Typography>

      <UsefulLinks
        problemId={problemId}
        problemInfo={problemInfoQuery.data}
        userId={auth?.state.user}
      />
      <Divider />

      <KatexRender text={problemInfoQuery.data.getStatement()} />

      <Divider
        sx={{
          margin: 1,
        }}
      />

      <Typography variant="h4" paragraph={true}>
        Submit
      </Typography>

      <form onSubmit={(e) => handleSubmit(e)}>
        <FormControl
          sx={{
            height: "400px",
            width: "100%",
          }}
        >
          <SourceEditor
            value={source}
            language={lang}
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
            value={lang}
            onChange={(e) => setLang(e.target.value as string)}
          >
            <MenuItem value="">Lang</MenuItem>
            {langListQuery.isSuccess &&
              langListQuery.data.getLangsList().map((e) => (
                <MenuItem key={e.getId()} value={e.getId()}>
                  {e.getName()}
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
export default ProblemInfo;
