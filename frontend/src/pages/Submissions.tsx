import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Paper from "@mui/material/Paper";
import TablePagination from "@mui/material/TablePagination";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import ListSubheader from "@mui/material/ListSubheader";
import React from "react";
import { useLocation } from "react-use";
import {
  useLangList,
  useProblemCategories,
  useProblemList,
  useSubmissionList,
} from "../api/client_wrapper";
import SubmissionTable from "../components/SubmissionTable";
import { categoriseProblems } from "../utils/problem.categorizer";
import { styled } from "@mui/system";
import KatexTypography from "../components/katex/KatexTypography";
import { Container } from "@mui/material";

const FilterFormControl = styled(FormControl)({
  margin: 1,
  verticalAlign: "bottom",
  minWidth: "120px",
});

const Submissions: React.FC = () => {
  const location = useLocation();
  const params = new URLSearchParams(location.search);

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

  const langListQuery = useLangList();

  const problemListQuery = useProblemList();
  const problemCategoriesQuery = useProblemCategories();

  const submissionListQuery = useSubmissionList(
    initialProblemName,
    initialUserName,
    initialStatusFilter,
    initialLangFilter,
    initialSortOrder,
    page * rowsPerPage,
    rowsPerPage
  );

  if (
    langListQuery.isLoading ||
    problemListQuery.isLoading ||
    problemCategoriesQuery.isLoading
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
    if (submissionListQuery.isLoading) {
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
        <SubmissionTable overviews={value.submissions} />
        <TablePagination
          rowsPerPage={rowsPerPage}
          component="div"
          count={value.count}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>
    );
  })();

  const categories = categoriseProblems(
    problemListQuery.data.problems,
    problemCategoriesQuery.data.categories
  );

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Submission List
      </Typography>
      <Box>
        <FilterFormControl variant="standard">
          <Select
            value={problemName}
            displayEmpty
            onChange={(e) => setProblemName(e.target.value as string)}
          >
            <MenuItem value="">Problem Name</MenuItem>
            {categories.map((category) =>
              [<ListSubheader>{category.name}</ListSubheader>].concat(
                category.problems.map((problem) => (
                  <MenuItem key={problem.name} value={problem.name}>
                    <KatexTypography>{problem.title}</KatexTypography>
                  </MenuItem>
                ))
              )
            )}
          </Select>
        </FilterFormControl>
        <FilterFormControl variant="standard">
          <TextField
            variant="standard"
            label="User"
            value={userName}
            autoComplete="off"
            onChange={(e) => setUserName(e.target.value)}
          />
        </FilterFormControl>
        <FilterFormControl variant="standard">
          <Select
            value={statusFilter}
            displayEmpty
            onChange={(e) => setStatusFilter(e.target.value as string)}
          >
            <MenuItem value="">Status</MenuItem>
            <MenuItem value="AC">AC</MenuItem>
          </Select>
        </FilterFormControl>
        <FilterFormControl variant="standard">
          <Select
            value={langFilter}
            displayEmpty
            onChange={(e) => setLangFilter(e.target.value as string)}
          >
            <MenuItem value="">Lang</MenuItem>
            {langListQuery.isSuccess &&
              langListQuery.data.langs.map((e) => (
                <MenuItem key={e.id} value={e.id}>
                  {e.name}
                </MenuItem>
              ))}
          </Select>
        </FilterFormControl>
        <FilterFormControl variant="standard">
          <Select
            value={sortOrder}
            displayEmpty
            onChange={(e) => setSortOrder(e.target.value as string)}
          >
            <MenuItem value="-id">Sort</MenuItem>
            <MenuItem value="+time">Time</MenuItem>
          </Select>
        </FilterFormControl>
        <Button
          variant="outlined"
          type="submit"
          href={`?${searchParams.toString()}`}
          onClick={() => submissionListQuery.remove()}
        >
          search
        </Button>
      </Box>

      {submissionList}
    </Container>
  );
};

export default Submissions;
