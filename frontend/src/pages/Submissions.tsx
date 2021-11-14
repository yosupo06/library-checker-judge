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
import { useQuery } from "react-query";
import { Link } from "react-router-dom";
import { useLocation } from "react-use";
import library_checker_client from "../api/library_checker_client";
import {
  LangListRequest,
  ProblemCategoriesRequest,
  ProblemListRequest,
  SubmissionListRequest,
} from "../api/library_checker_pb";
import KatexRender from "../components/KatexRender";
import SubmissionTable from "../components/SubmissionTable";
import { getCategories } from "../utils/ProblemCategory";

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

const Submissions: React.FC = () => {
  const params = new URLSearchParams(useLocation().search);
  const classes = useStyles();

  const initialProblemName = params.get("problem") ?? "";
  const [problemName, setProblemName] = React.useState(initialProblemName);
  const initialUserName = params.get("user") ?? "";
  const [userName, setUserName] = React.useState(initialUserName);
  const initialStatusFilter = params.get("status") ?? "";
  const [statusFilter, setStatusFilter] = React.useState(initialStatusFilter);
  const initialLangFilter = params.get("lang") ?? "";
  const [langFilter, setLangFilter] = React.useState(initialLangFilter);
  const initialSortOrder = params.get("order") ?? "-id";
  const [sortOrder, setSortOrder] = React.useState(initialSortOrder);

  const [page, setPage] = React.useState(parseInt(params.get("page") ?? "0"));
  const [rowsPerPage, setRowsPerPage] = React.useState(
    parseInt(params.get("pagesize") ?? "100")
  );

  const searchParams = new URLSearchParams({
    problem: problemName,
    user: userName,
    status: statusFilter,
    lang: langFilter,
    order: sortOrder,
    page: page.toString(),
    pagesize: rowsPerPage.toString(),
  });

  const langListQuery = useQuery("langList", () =>
    library_checker_client.langList(new LangListRequest(), {})
  );
  const problemListQuery = useQuery("problemList", () =>
    library_checker_client.problemList(new ProblemListRequest(), {})
  );
  const problemCategoriesQuery = useQuery("problemCategories", () =>
    library_checker_client.problemCategories(new ProblemCategoriesRequest(), {})
  );
  const submissionListQuery = useQuery(
    [
      "submissionList",
      initialProblemName,
      initialUserName,
      initialStatusFilter,
      initialLangFilter,
      initialSortOrder,
      page,
      rowsPerPage,
    ],
    () =>
      library_checker_client.submissionList(
        new SubmissionListRequest()
          .setProblem(initialProblemName)
          .setUser(initialUserName)
          .setStatus(initialStatusFilter)
          .setLang(initialLangFilter)
          .setOrder(initialSortOrder)
          .setSkip(page * rowsPerPage)
          .setLimit(rowsPerPage),
        {}
      )
  );

  if (
    langListQuery.isLoading ||
    langListQuery.isIdle ||
    problemListQuery.isLoading ||
    problemListQuery.isIdle ||
    problemCategoriesQuery.isLoading ||
    problemCategoriesQuery.isIdle
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
    langListQuery.isError ||
    problemListQuery.isError ||
    problemCategoriesQuery.isError
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
    if (submissionListQuery.isLoading || submissionListQuery.isIdle) {
      return (
        <Paper>
          <SubmissionTable overviews={[]} />
          <CircularProgress />
        </Paper>
      );
    }
    if (submissionListQuery.isError) {
      return <p>Error: {submissionListQuery.error}</p>;
    }
    const value = submissionListQuery.data;
    return (
      <Paper>
        <SubmissionTable overviews={value.getSubmissionsList()} />
        <TablePagination
          rowsPerPage={rowsPerPage}
          component="div"
          count={value.getCount()}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>
    );
  })();

  const categories = getCategories(
    problemListQuery.data.getProblemsList(),
    problemCategoriesQuery.data.getCategoriesList()
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
            {langListQuery.isSuccess &&
              langListQuery.data.getLangsList().map((e) => (
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
            to={{
              search: searchParams.toString(),
            }}
            onClick={() => submissionListQuery.remove()}
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

export default Submissions;
