import {
  Box,
  Button,
  CircularProgress,
  FormControl,
  ListSubheader,
  makeStyles,
  MenuItem,
  Paper,
  Select,
  TablePagination,
  TextField,
  Theme,
  Typography,
} from "@material-ui/core";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import { Link } from "react-router-dom";
import { useLocation } from "react-use";
import library_checker_client from "../api/library_checker_client";
import {
  LangListRequest,
  LangListResponse,
  ProblemCategoriesRequest,
  ProblemCategoriesResponse,
  ProblemListRequest,
  ProblemListResponse,
  SubmissionListRequest,
  SubmissionListResponse,
} from "../api/library_checker_pb";
import KatexRender from "../components/KatexRender";
import SubmissionTable from "../components/SubmissionTable";
import { getCategories } from "../utils/ProblemCategory";

interface OuterProps {}

interface BridgeProps {
  problem: string;
  user: string;
  status: string;
  lang: string;
  order: string;
  page: number;
  pageSize: number;
}

interface InnerProps extends BridgeProps {
  langListFetch: PromiseState<LangListResponse>;
  problemListFetch: PromiseState<ProblemListResponse>;
  problemCategoriesFetch: PromiseState<ProblemCategoriesResponse>;
  submissionListFetch: PromiseState<SubmissionListResponse>;
}

const useStyles = makeStyles((theme: Theme) => ({
  formControl: {
    margin: theme.spacing(1),
    verticalAlign: "bottom",
    minWidth: 120,
  },
  searchLink: {
    color: "inherit",
    textDecoration: "none",
  },
}));

const InnerSubmissions: React.FC<InnerProps> = (props) => {
  const {
    langListFetch,
    problemListFetch,
    problemCategoriesFetch,
    submissionListFetch,
  } = props;
  const classes = useStyles();
  const [problemName, setProblemName] = React.useState(props.problem);
  const [userName, setUserName] = React.useState(props.user);
  const [statusFilter, setStatusFilter] = React.useState(props.status);
  const [langFilter, setLangFilter] = React.useState(props.lang);
  const [sortOrder, setSortOrder] = React.useState(props.order);
  const [page, setPage] = React.useState(props.page);
  const [rowsPerPage, setRowsPerPage] = React.useState(props.pageSize);

  const searchParams = new URLSearchParams({
    problem: problemName,
    user: userName,
    status: statusFilter,
    lang: langFilter,
    order: sortOrder,
    page: page.toString(),
    pageSize: rowsPerPage.toString(),
  });

  if (
    langListFetch.pending ||
    problemListFetch.pending ||
    problemCategoriesFetch.pending
  ) {
    return (
      <Box>
        <Typography variant="h2" paragraph={true}>
          Submission List
        </Typography>
        <CircularProgress />
      </Box>
    );
  }
  if (
    langListFetch.rejected ||
    problemListFetch.rejected ||
    problemCategoriesFetch.rejected
  ) {
    return (
      <Box>
        <Typography variant="h2" paragraph={true}>
          Submission List
        </Typography>
        Error
      </Box>
    );
  }

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

  const categories = getCategories(
    problemListFetch.value.getProblemsList(),
    problemCategoriesFetch.value.getCategoriesList()
  );

  return (
    <Box>
      <Typography variant="h2" paragraph={true}>
        Submission List
      </Typography>
      <Box>
        <FormControl className={classes.formControl}>
          <Select
            value={problemName}
            displayEmpty
            onChange={(e) => setProblemName(e.target.value as string)}
          >
            <MenuItem value="">Problem Name</MenuItem>
            {categories.map((category) =>
              [<ListSubheader>{category.name}</ListSubheader>].concat(
                category.problems.map((e) => (
                  <MenuItem key={e.name} value={e.name}>
                    <KatexRender text={e.title} />
                  </MenuItem>
                ))
              )
            )}
          </Select>
        </FormControl>
        <FormControl className={classes.formControl}>
          <TextField
            label="User Name"
            value={userName}
            autoComplete="off"
            onChange={(e) => setUserName(e.target.value)}
          />
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
            {langListFetch.fulfilled &&
              langListFetch.value.getLangsList().map((e) => (
                <MenuItem key={e.getId()} value={e.getId()}>
                  {e.getName()}
                </MenuItem>
              ))}
          </Select>
        </FormControl>
        <FormControl className={classes.formControl}>
          <Select
            value={sortOrder}
            displayEmpty
            onChange={(e) => setSortOrder(e.target.value as string)}
          >
            <MenuItem value="-id">Sort</MenuItem>
            <MenuItem value="+time">Time</MenuItem>
          </Select>
        </FormControl>
        <Button variant="outlined" type="submit">
          <Link
            to={`/submissions?${searchParams.toString()}`}
            className={classes.searchLink}
          >
            search
          </Link>
        </Button>
      </Box>

      {submissionList}
    </Box>
  );
};

const BridgeSubmissions = connect<BridgeProps, InnerProps>((props) => ({
  langListFetch: {
    comparison: null,
    value: () => library_checker_client.langList(new LangListRequest(), {}),
  },
  problemListFetch: {
    comparison: null,
    value: () =>
      library_checker_client.problemList(new ProblemListRequest(), {}),
  },
  problemCategoriesFetch: {
    comparison: null,
    value: () =>
      library_checker_client.problemCategories(
        new ProblemCategoriesRequest(),
        {}
      ),
  },
  submissionListFetch: {
    comparison: `${props.problem}/${props.user}/${props.status}/${props.lang}/${props.order}/${props.page}/${props.pageSize}`,
    value: () =>
      library_checker_client.submissionList(
        new SubmissionListRequest()
          .setProblem(props.problem)
          .setUser(props.user)
          .setStatus(props.status)
          .setLang(props.lang)
          .setOrder(props.order)
          .setSkip(props.page * props.pageSize)
          .setLimit(props.pageSize),
        {}
      ),
  },
}))(InnerSubmissions);

const Submissions: React.FC<OuterProps> = (props: OuterProps) => {
  const params = new URLSearchParams(useLocation().search);
  console.log("Yosupo");
  return (
    <BridgeSubmissions
      user={params.get("user") ?? ""}
      problem={params.get("problem") ?? ""}
      status={params.get("status") ?? ""}
      lang={params.get("lang") ?? ""}
      order={params.get("order") ?? "-id"}
      page={parseInt(params.get("page") ?? "0")}
      pageSize={parseInt(params.get("page_size") ?? "100")}
    />
  );
};

export default Submissions;
