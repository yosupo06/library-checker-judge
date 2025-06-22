import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import TextField from "@mui/material/TextField";
import Paper from "@mui/material/Paper";
import TablePagination from "@mui/material/TablePagination";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import ListSubheader from "@mui/material/ListSubheader";
import React from "react";
import { useLocation } from "react-use";
import CachedIcon from "@mui/icons-material/Cached";
import {
  useLangList,
  useProblemCategories,
  useProblemList,
  useSubmissionList,
} from "../api/client_wrapper";
import SubmissionTable from "../components/SubmissionTable";
import { categoriseProblems } from "../utils/problem.categorizer";
import { styled } from "@mui/material/styles";
import KatexTypography from "../components/katex/KatexTypography";
import { Checkbox, FormControlLabel, InputLabel } from "@mui/material";
import MainContainer from "../components/MainContainer";

type SearchParams = {
  problem: string;
  user: string;
  dedupUser: boolean;
  status: string;
  lang: string;
  order: string;
  page: number;
  rowsPerPage: number;
};

const toURLSearchParams = (searchParams: SearchParams) => {
  return new URLSearchParams({
    problem: searchParams.problem,
    user: searchParams.user,
    dedupUser: searchParams.dedupUser.toString(),
    status: searchParams.status,
    lang: searchParams.lang,
    order: searchParams.order,
    page: searchParams.page.toString(),
    pagesize: searchParams.rowsPerPage.toString(),
  });
};

const Submissions: React.FC = () => {
  const location = useLocation();
  const params = new URLSearchParams(location.search);

  const [searchParams, setSearchParams] = React.useState({
    problem: params.get("problem") ?? "",
    user: params.get("user") ?? "",
    dedupUser: (params.get("dedupuser") ?? "") === "true",
    status: params.get("status") ?? "",
    lang: params.get("lang") ?? "",
    order: params.get("order") ?? "-id",
    page: parseInt(params.get("page") ?? "0"),
    rowsPerPage: parseInt(params.get("pagesize") ?? "100"),
  });

  return (
    <MainContainer title="Submission List">
      <SubmissionsForm
        searchParams={searchParams}
        setSearchParams={setSearchParams}
      />
      <SubmissionsBody
        searchParams={searchParams}
        setSearchParams={setSearchParams}
      />
    </MainContainer>
  );
};

const FilterFormControl = styled(FormControl)({
  margin: 1,
  verticalAlign: "bottom",
  minWidth: "120px",
});

const SubmissionsForm: React.FC<{
  searchParams: SearchParams;
  setSearchParams: (params: SearchParams) => void;
}> = (props) => {
  const { searchParams, setSearchParams } = props;

  const langListQuery = useLangList();
  const problemListQuery = useProblemList();
  const problemCategoriesQuery = useProblemCategories();

  if (
    langListQuery.isPending ||
    problemListQuery.isPending ||
    problemCategoriesQuery.isPending
  ) {
    return (
      <Box>
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
      // TODO: use <Alert>
      <Box>Error</Box>
    );
  }

  const categories = categoriseProblems(
    problemListQuery.data.problems,
    problemCategoriesQuery.data.categories,
  );

  return (
    <Box>
      <FilterFormControl variant="standard">
        <InputLabel>Problem</InputLabel>
        <Select
          value={searchParams.problem}
          onChange={(e) =>
            setSearchParams({
              ...searchParams,
              problem: e.target.value,
            })
          }
        >
          <MenuItem value="">-</MenuItem>
          {categories.map((category) =>
            [<ListSubheader>{category.name}</ListSubheader>].concat(
              category.problems.map((problem) => (
                <MenuItem key={problem.name} value={problem.name}>
                  <KatexTypography>{problem.title}</KatexTypography>
                </MenuItem>
              )),
            ),
          )}
        </Select>
      </FilterFormControl>
      <FilterFormControl variant="standard">
        <TextField
          variant="standard"
          label="User"
          value={searchParams.user}
          autoComplete="off"
          onChange={(e) =>
            setSearchParams({
              ...searchParams,
              user: e.target.value,
              page: 0,
            })
          }
        />
      </FilterFormControl>
      <FilterFormControl variant="standard">
        <InputLabel>Status</InputLabel>
        <Select
          value={searchParams.status}
          onChange={(e) =>
            setSearchParams({
              ...searchParams,
              status: e.target.value,
              page: 0,
            })
          }
        >
          <MenuItem value="">-</MenuItem>
          <MenuItem value="AC">AC</MenuItem>
        </Select>
      </FilterFormControl>
      <FilterFormControl variant="standard">
        <InputLabel>Language</InputLabel>
        <Select
          value={searchParams.lang}
          onChange={(e) =>
            setSearchParams({
              ...searchParams,
              lang: e.target.value,
              page: 0,
            })
          }
        >
          <MenuItem value="">-</MenuItem>
          {langListQuery.isSuccess &&
            langListQuery.data.langs.map((e) => (
              <MenuItem key={e.id} value={e.id}>
                {e.name}
              </MenuItem>
            ))}
        </Select>
      </FilterFormControl>
      <FilterFormControl variant="standard">
        <InputLabel>Sort</InputLabel>
        <Select
          value={searchParams.order}
          displayEmpty
          onChange={(e) =>
            setSearchParams({
              ...searchParams,
              order: e.target.value,
              page: 0,
            })
          }
        >
          <MenuItem value="-id">ID</MenuItem>
          <MenuItem value="+time">Time</MenuItem>
        </Select>
      </FilterFormControl>
      <Button
        variant="outlined"
        type="submit"
        href={`?${toURLSearchParams(searchParams).toString()}`}
      >
        <CachedIcon></CachedIcon>
      </Button>
      <FormControlLabel
        control={<Checkbox />}
        label="Dedup user"
        checked={searchParams.dedupUser}
        onChange={(_e, checked) =>
          setSearchParams({
            ...searchParams,
            dedupUser: checked,
            page: 0,
          })
        }
      />
    </Box>
  );
};

const SubmissionsBody: React.FC<{
  searchParams: SearchParams;
  setSearchParams: (params: SearchParams) => void;
}> = (props) => {
  const { searchParams, setSearchParams } = props;
  const { problem, user, dedupUser, status, lang, order, page, rowsPerPage } =
    searchParams;

  const submissionListQuery = useSubmissionList(
    problem,
    user,
    dedupUser,
    status,
    lang,
    order,
    page * rowsPerPage,
    rowsPerPage,
  );

  if (submissionListQuery.isPending) {
    return (
      <Paper>
        <SubmissionTable overviews={[]} />
        <CircularProgress />
      </Paper>
    );
  }
  if (submissionListQuery.isError) {
    return <Box>Error: {submissionListQuery.error.message}</Box>;
  }

  const handleChangePage = (
    _: React.MouseEvent<HTMLButtonElement> | null,
    newPage: number,
  ) => {
    setSearchParams({
      ...searchParams,
      page: newPage,
    });
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    setSearchParams({
      ...searchParams,
      rowsPerPage: parseInt(event.target.value, 10),
    });
  };

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
};

export default Submissions;
