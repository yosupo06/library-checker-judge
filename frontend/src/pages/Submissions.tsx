import {
  Button,
  CircularProgress,
  Container,
  createStyles,
  FormControl,
  makeStyles,
  MenuItem,
  Paper,
  Select,
  TablePagination,
  TextField,
  Theme,
  Typography
} from "@material-ui/core";
import React, { useEffect } from "react";
import { connect, PromiseState } from "react-refetch";
import library_checker_client from "../api/library_checker_client";
import {
  LangListRequest,
  LangListResponse,
  ProblemListRequest,
  ProblemListResponse,
  SubmissionListRequest,
  SubmissionListResponse
} from "../api/library_checker_pb";
import KatexRender from "../components/KatexRender";
import SubmissionTable from "../components/SubmissionTable";

interface Props {
  langListFetch: PromiseState<LangListResponse>;
  problemListFetch: PromiseState<ProblemListResponse>;
  submissionListFetch: PromiseState<SubmissionListResponse>;
  refreshSubmissionList: (
    problem: string,
    user: string,
    status: string,
    lang: string,
    skip: number,
    limit: number
  ) => void;
}

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    formControl: {
      margin: theme.spacing(1),
      verticalAlign: "bottom",
      minWidth: 120
    }
  })
);

const Submissions: React.FC<Props> = props => {
  const {
    langListFetch,
    problemListFetch,
    submissionListFetch,
    refreshSubmissionList
  } = props;
  const classes = useStyles();
  const [queryProblemName, setQueryProblemName] = React.useState("");
  const [problemName, setProblemName] = React.useState("");
  const [queryUserName, setQueryUserName] = React.useState("");
  const [userName, setUserName] = React.useState("");
  const [queryStatus, setQueryStatus] = React.useState("");
  const [statusFilter, setStatusFilter] = React.useState("");
  const [queryLang, setQueryLang] = React.useState("");
  const [langFilter, setLangFilter] = React.useState("");
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(100);

  useEffect(
    () =>
      refreshSubmissionList(
        queryProblemName,
        queryUserName,
        queryStatus,
        queryLang,
        page * rowsPerPage,
        rowsPerPage
      ),
    [
      refreshSubmissionList,
      queryProblemName,
      queryUserName,
      queryStatus,
      queryLang,
      page,
      rowsPerPage
    ]
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setQueryProblemName(problemName);
    setQueryUserName(userName);
    setQueryStatus(statusFilter);
    setQueryLang(langFilter);
    setPage(0);
  };

  const handleChangePage = (
    _: React.MouseEvent<HTMLButtonElement> | null,
    newPage: number
  ) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const newRowsPerPage = parseInt(event.target.value, 10);
    setRowsPerPage(newRowsPerPage);
    setPage(0);
  };

  const submissionList = (() => {
    if (submissionListFetch.pending) {
      return (
        <Paper>
          <SubmissionTable overviews={[]} />
          <CircularProgress />
        </Paper>
      );
    }
    if (submissionListFetch.rejected) {
      return <p>{submissionListFetch.reason}</p>;
    }
    const value = submissionListFetch.value;
    return (
      <Paper>
        <SubmissionTable overviews={value.getSubmissionsList()} />
        <TablePagination
          rowsPerPage={rowsPerPage}
          component="div"
          count={value.getCount()}
          page={page}
          onChangePage={handleChangePage}
          onChangeRowsPerPage={handleChangeRowsPerPage}
        />
      </Paper>
    );
  })();

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Submission List
      </Typography>
      <form onSubmit={e => handleSubmit(e)}>
        <FormControl className={classes.formControl}>
          <Select
            value={problemName}
            displayEmpty
            onChange={e => setProblemName(e.target.value as string)}
          >
            <MenuItem value="">Problem Name</MenuItem>
            {problemListFetch.fulfilled &&
              problemListFetch.value.getProblemsList().map(e => (
                <MenuItem key={e.getName()} value={e.getName()}>
                  <KatexRender text={e.getTitle()} />
                </MenuItem>
              ))}
          </Select>
        </FormControl>
        <FormControl className={classes.formControl}>
          <TextField
            label="User Name"
            value={userName}
            onChange={e => setUserName(e.target.value)}
          />
        </FormControl>
        <FormControl className={classes.formControl}>
          <Select
            value={statusFilter}
            displayEmpty
            onChange={e => setStatusFilter(e.target.value as string)}
          >
            <MenuItem value="">Status</MenuItem>
            <MenuItem value="AC">AC</MenuItem>
          </Select>
        </FormControl>
        <FormControl className={classes.formControl}>
          <Select
            value={langFilter}
            displayEmpty
            onChange={e => setLangFilter(e.target.value as string)}
          >
            <MenuItem value="">Lang</MenuItem>
            {langListFetch.fulfilled &&
              langListFetch.value.getLangsList().map(e => (
                <MenuItem key={e.getId()} value={e.getId()}>
                  {e.getName()}
                </MenuItem>
              ))}
          </Select>
        </FormControl>
        <Button color="primary" type="submit">
          Search
        </Button>
      </form>

      {submissionList}
    </Container>
  );
};

export default connect<{}, Props>(() => ({
  langListFetch: {
    comparison: null,
    value: () => library_checker_client.langList(new LangListRequest())
  },
  problemListFetch: {
    comparison: null,
    value: () => library_checker_client.problemList(new ProblemListRequest())
  },
  submissionListFetch: {
    comparison: null,
    value: []
  },
  refreshSubmissionList: (
    problem: string,
    user: string,
    status: string,
    lang: string,
    skip: number,
    limit: number
  ) => ({
    submissionListFetch: {
      comparison: `${problem}:${user}:${status}:${lang}:${skip}:${limit}`,
      refreshing: true,
      value: () =>
        library_checker_client.submissionList(
          new SubmissionListRequest()
            .setProblem(problem)
            .setUser(user)
            .setStatus(status)
            .setLang(lang)
            .setSkip(skip)
            .setLimit(limit)
        )
    }
  })
}))(Submissions);
