import {
  Box,
  Button,
  CircularProgress,
  Divider,
  FormControl,
  makeStyles,
  MenuItem,
  Select,
  Typography,
} from "@material-ui/core";
import React, { useContext, useState } from "react";
import { connect, PromiseState } from "react-refetch";
import { RouteComponentProps } from "react-router-dom";
import { useLocalStorage } from "react-use";
import library_checker_client, {
  authMetadata,
} from "../api/library_checker_client";
import {
  LangListRequest,
  LangListResponse,
  ProblemInfoRequest,
  ProblemInfoResponse,
  SubmitRequest,
} from "../api/library_checker_pb";
import Editor from "../components/Editor";
import KatexRender from "../components/KatexRender";
import { AuthContext } from "../contexts/AuthContext";

interface Props extends RouteComponentProps<{ problemId: string }> {
  problemInfoFetch: PromiseState<ProblemInfoResponse>;
  langListFetch: PromiseState<LangListResponse>;
}

const useStyles = makeStyles((theme) => ({
  divider: {
    margin: theme.spacing(1),
  },
  editor: {
    height: "400px",
    width: "100%",
  },
}));

const ProblemInfo: React.FC<Props> = (props) => {
  const classes = useStyles();
  const { problemInfoFetch, history } = props;
  const auth = useContext(AuthContext);
  const [source, setSource] = useState("");
  const [lang, setLang] = useLocalStorage("programming-lang", "");

  if (problemInfoFetch.pending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (problemInfoFetch.rejected) {
    return (
      <Box>
        <Typography variant="body1">Error</Typography>
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
  return (
    <Box>
      <Typography variant="h2" paragraph={true}>
        <KatexRender text={problemInfoFetch.value.getTitle()} />
      </Typography>
      <Typography variant="body1" paragraph={true}>
        Time Limit: {problemInfoFetch.value.getTimeLimit()} sec
      </Typography>

      <KatexRender text={problemInfoFetch.value.getStatement()} html={true} />

      <Divider className={classes.divider} />

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
            {props.langListFetch.fulfilled &&
              props.langListFetch.value.getLangsList().map((e) => (
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

export default connect<RouteComponentProps<{ problemId: string }>, Props>(
  (props) => ({
    problemInfoFetch: {
      comparison: null,
      value: () =>
        library_checker_client.problemInfo(
          new ProblemInfoRequest().setName(props.match.params.problemId),
          {}
        ),
    },
    langListFetch: {
      comparison: null,
      value: () => library_checker_client.langList(new LangListRequest(), {}),
    },
  })
)(ProblemInfo);
