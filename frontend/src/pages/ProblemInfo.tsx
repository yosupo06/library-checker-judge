import {
  Box,
  Button,
  CircularProgress,
  Container,
  FormControl,
  MenuItem,
  Select,
  TextField,
  Typography
} from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import React, { useContext, useState } from "react";
import { connect, PromiseState } from "react-refetch";
import { Link, RouteComponentProps } from "react-router-dom";
import library_checker_client, {
  authMetadata
} from "../api/library_checker_client";
import {
  LangListRequest,
  LangListResponse,
  ProblemInfoRequest,
  ProblemInfoResponse,
  SubmitRequest
} from "../api/library_checker_pb";
import KatexRender from "../components/KatexRender";
import { AuthContext } from "../contexts/AuthContext";

interface Props extends RouteComponentProps<{ problemId: string }> {
  problemInfoFetch: PromiseState<ProblemInfoResponse>;
  langListFetch: PromiseState<LangListResponse>;
}

const ProblemInfo: React.FC<Props> = props => {
  const { problemInfoFetch, history } = props;
  const auth = useContext(AuthContext);
  const [source, setSource] = useState("");
  const [lang, setLang] = useState("");

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
    console.log(props);
    library_checker_client
      .submit(
        new SubmitRequest()
          .setLang(lang)
          .setProblem(props.match.params.problemId)
          .setSource(source),
        authMetadata(auth?.state!)
      )
      .then(resp => {
        history.push(`/submission/${resp.getId()}`);
      });
  };
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        <KatexRender text={problemInfoFetch.value.getTitle()} />
      </Typography>
      <Typography variant="body1" paragraph={true}>
        Time Limit: {problemInfoFetch.value.getTimeLimit()} sec
      </Typography>

      <KatexRender text={problemInfoFetch.value.getStatement()} html={true} />

      <form onSubmit={e => handleSubmit(e)}>
        <FormControl>
          <TextField
            required
            multiline
            label="source"
            value={source}
            rows={10}
            onChange={e => setSource(e.target.value)}
          />
        </FormControl>
        <FormControl>
          <Select
            displayEmpty
            required
            value={lang}
            onChange={e => setLang(e.target.value as string)}
          >
            <MenuItem value="">Lang</MenuItem>
            {props.langListFetch.fulfilled &&
              props.langListFetch.value.getLangsList().map(e => (
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
    </Container>
  );
};

export default connect<RouteComponentProps<{ problemId: string }>, Props>(
  props => ({
    problemInfoFetch: {
      comparison: null,
      value: () =>
        library_checker_client.problemInfo(
          new ProblemInfoRequest().setName(props.match.params.problemId)
        )
    },
    langListFetch: {
      comparison: null,
      value: () => library_checker_client.langList(new LangListRequest())
    }
  })
)(ProblemInfo);
