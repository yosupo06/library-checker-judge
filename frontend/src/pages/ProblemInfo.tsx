import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import { makeStyles } from "@material-ui/core";
import React, { useContext, useState } from "react";
import { Link, RouteComponentProps } from "react-router-dom";
import { useLocalStorage } from "react-use";
import library_checker_client, {
  authMetadata,
  useLangList,
  useProblemInfo,
} from "../api/library_checker_client";
import { SubmitRequest } from "../api/library_checker_pb";
import Editor from "../components/Editor";
import KatexRender from "../components/KatexRender";
import { AuthContext } from "../contexts/AuthContext";
import { GitHub, FlashOn } from "@mui/icons-material";

const useStyles = makeStyles((theme) => ({
  divider: {
    margin: theme.spacing(1),
  },
  editor: {
    height: "400px",
    width: "100%",
  },
  button: {
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
  fastestLink: {
    color: "inherit",
    textDecoration: "none",
  },
}));

const ProblemInfo: React.FC<RouteComponentProps<{ problemId: string }>> = (
  props
) => {
  const classes = useStyles();
  const { history } = props;
  const auth = useContext(AuthContext);
  const [source, setSource] = useState("");
  const [lang, setLang] = useLocalStorage("programming-lang", "");

  const problemInfoQuery = useProblemInfo(props.match.params.problemId);
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
          .setProblem(props.match.params.problemId)
          .setSource(source),
        (auth && authMetadata(auth.state)) ?? null
      )
      .then((resp) => {
        history.push(`/submission/${resp.getId()}`);
      });
  };
  const fastestParams = new URLSearchParams({
    problem: props.match.params.problemId,
    order: "+time",
    status: "AC",
  });
  return (
    <Box>
      <Typography variant="h2" paragraph={true}>
        <KatexRender text={problemInfoQuery.data.getTitle()} />
      </Typography>
      <Typography variant="body1" paragraph={true}>
        Time Limit: {problemInfoQuery.data.getTimeLimit()} sec
      </Typography>
      <Button
        variant="contained"
        color="info"
        className={classes.button}
        startIcon={<FlashOn />}
      >
        <Link
          to={`/submissions/?${fastestParams.toString()}`}
          className={classes.fastestLink}
        >
          Fastest
        </Link>
      </Button>
      <Button
        variant="contained"
        color="info"
        className={classes.button}
        startIcon={<GitHub />}
        href={problemInfoQuery.data.getSourceUrl()}
      >
        Github
      </Button>
      <Divider className={classes.divider} />

      <KatexRender text={problemInfoQuery.data.getStatement()} html={true} />

      <form onSubmit={(e) => handleSubmit(e)}>
        <FormControl className={classes.editor}>
          <Editor
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
