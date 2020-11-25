import { Button, Container, createStyles, FormControl, makeStyles, MenuItem, Select, TextField, Theme, Typography } from '@material-ui/core';
import React from 'react';
import { connect, PromiseState } from 'react-refetch';
import library_checker_client from '../api/library_checker_client';
import { LangListRequest, LangListResponse, ProblemListRequest, ProblemListResponse, SubmissionListRequest, SubmissionOverview } from "../api/library_checker_pb";
import KatexRender from '../components/KatexRender';
import SubmissionList from '../components/SubmissionList';

interface Props {
  langListFetch: PromiseState<LangListResponse>;
  problemListFetch: PromiseState<ProblemListResponse>;
}

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    formControl: {
      margin: theme.spacing(1),
      verticalAlign: 'bottom',
      minWidth: 120,
    },
  }),
);

const Submissions: React.FC<Props> = (props) => {
  const classes = useStyles()
  const [problemName, setProblemName] = React.useState("")
  const [userName, setUserName] = React.useState("")
  const [statusFilter, setStatusFilter] = React.useState("")
  const [langFilter, setLangFilter] = React.useState("")
  const [submissionOverviews, setSubmissionOverviews] = React.useState<SubmissionOverview[]>([])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const request = new SubmissionListRequest()
      .setUser(userName)
      .setProblem(problemName)
      .setStatus(statusFilter)
      .setLang(langFilter)
      .setLimit(100)
    library_checker_client.submissionList(request)
      .then((resp) => {
        setSubmissionOverviews(resp.getSubmissionsList())
      })
  }
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>Submission List</Typography>
      <form onSubmit={e => handleSubmit(e)}>
        <FormControl className={classes.formControl}>
          <Select
            value={problemName}
            displayEmpty
            onChange={(e) => setProblemName(e.target.value as string)}
          >
            <MenuItem value="">Problem Name</MenuItem>
            {props.problemListFetch.fulfilled && props.problemListFetch.value.getProblemsList().map(e =>
              <MenuItem key={e.getName()} value={e.getName()}><KatexRender text={e.getTitle()} /></MenuItem>
            )}
          </Select>
        </FormControl>
        <FormControl className={classes.formControl}>
          <TextField label="User Name" value={userName} onChange={(e) => setUserName(e.target.value)} />
        </FormControl>
        <FormControl className={classes.formControl}>
          <Select
            value={statusFilter}
            displayEmpty
            onChange={(e) => setStatusFilter(e.target.value as string)}
          >
            <MenuItem value="">Status</MenuItem>
            <MenuItem value="AC">AC</MenuItem>
          </Select>
        </FormControl>
        <FormControl className={classes.formControl}>
          <Select
            value={langFilter}
            displayEmpty
            onChange={(e) => setLangFilter(e.target.value as string)}
          >
            <MenuItem value="">Lang</MenuItem>
            {props.langListFetch.fulfilled && props.langListFetch.value.getLangsList().map(e =>
              <MenuItem key={e.getId()} value={e.getId()}>{e.getName()}</MenuItem>
            )}
          </Select>
        </FormControl>
        <Button color="primary" type="submit">Search</Button>
      </form>

      <SubmissionList submissionOverviews={submissionOverviews} />
    </Container>
  );
}

export default connect<{}, Props>(() => ({
  langListFetch: {
    comparison: null,
    value: library_checker_client.langList(new LangListRequest())
  },
  problemListFetch: {
    comparison: null,
    value: library_checker_client.problemList(new ProblemListRequest())
  }
}))(Submissions);
