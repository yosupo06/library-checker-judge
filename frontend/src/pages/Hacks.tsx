import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import TextField from "@mui/material/TextField";
import Paper from "@mui/material/Paper";
import TablePagination from "@mui/material/TablePagination";
import FormControl from "@mui/material/FormControl";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import React from "react";
import { useLocation } from "react-use";
import CachedIcon from "@mui/icons-material/Cached";
import { useHackList } from "../api/client_wrapper";
import HackTable from "../components/HackTable";
import { styled } from "@mui/material/styles";
import { InputLabel } from "@mui/material";
import MainContainer from "../components/MainContainer";

type SearchParams = {
  user: string;
  status: string;
  order: string;
  page: number;
  rowsPerPage: number;
};

const toURLSearchParams = (searchParams: SearchParams) => {
  return new URLSearchParams({
    user: searchParams.user,
    status: searchParams.status,
    order: searchParams.order,
    page: searchParams.page.toString(),
    pagesize: searchParams.rowsPerPage.toString(),
  });
};

const Hacks: React.FC = () => {
  const location = useLocation();
  const params = new URLSearchParams(location.search);

  const [searchParams, setSearchParams] = React.useState({
    user: params.get("user") ?? "",
    status: params.get("status") ?? "",
    order: params.get("order") ?? "-id",
    page: parseInt(params.get("page") ?? "0"),
    rowsPerPage: parseInt(params.get("pagesize") ?? "100"),
  });

  return (
    <MainContainer title="Hack List">
      <HacksForm
        searchParams={searchParams}
        setSearchParams={setSearchParams}
      />
      <HacksBody
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

const HacksForm: React.FC<{
  searchParams: SearchParams;
  setSearchParams: (params: SearchParams) => void;
}> = (props) => {
  const { searchParams, setSearchParams } = props;

  const handleSearch = () => {
    const params = toURLSearchParams(searchParams);
    window.history.replaceState(null, "", "?" + params.toString());
    setSearchParams({ ...searchParams, page: 0 });
  };

  return (
    <Paper sx={{ p: 2, my: 2 }}>
      <Box
        sx={{ display: "flex", flexWrap: "wrap", gap: 1, alignItems: "end" }}
      >
        <TextField
          label="User"
          value={searchParams.user}
          onChange={(e) =>
            setSearchParams({ ...searchParams, user: e.target.value })
          }
          variant="outlined"
          size="small"
        />

        <FilterFormControl>
          <InputLabel>Status</InputLabel>
          <Select
            value={searchParams.status}
            onChange={(e) =>
              setSearchParams({ ...searchParams, status: e.target.value })
            }
            label="Status"
            size="small"
          >
            <MenuItem value="">All</MenuItem>
            <MenuItem value="AC">AC</MenuItem>
            <MenuItem value="WA">WA</MenuItem>
            <MenuItem value="TLE">TLE</MenuItem>
            <MenuItem value="MLE">MLE</MenuItem>
            <MenuItem value="RE">RE</MenuItem>
            <MenuItem value="CE">CE</MenuItem>
            <MenuItem value="WJ">WJ</MenuItem>
          </Select>
        </FilterFormControl>

        <FilterFormControl>
          <InputLabel>Order</InputLabel>
          <Select
            value={searchParams.order}
            onChange={(e) =>
              setSearchParams({ ...searchParams, order: e.target.value })
            }
            label="Order"
            size="small"
          >
            <MenuItem value="-id">ID (Desc)</MenuItem>
            <MenuItem value="id">ID (Asc)</MenuItem>
            <MenuItem value="-time">Time (Desc)</MenuItem>
            <MenuItem value="time">Time (Asc)</MenuItem>
          </Select>
        </FilterFormControl>

        <Button
          onClick={handleSearch}
          variant="contained"
          startIcon={<CachedIcon />}
        >
          Search
        </Button>
      </Box>
    </Paper>
  );
};

const HacksBody: React.FC<{
  searchParams: SearchParams;
  setSearchParams: (params: SearchParams) => void;
}> = (props) => {
  const { searchParams, setSearchParams } = props;

  const hackListQuery = useHackList(
    searchParams.user,
    searchParams.status,
    searchParams.order,
    searchParams.page * searchParams.rowsPerPage,
    searchParams.rowsPerPage,
  );

  if (hackListQuery.isPending) {
    return <CircularProgress />;
  }

  if (hackListQuery.isError) {
    return <div>Error loading hacks</div>;
  }

  const hackList = hackListQuery.data;

  const handleChangePage = (
    _event: React.MouseEvent<HTMLButtonElement> | null,
    newPage: number,
  ) => {
    const newSearchParams = { ...searchParams, page: newPage };
    const params = toURLSearchParams(newSearchParams);
    window.history.replaceState(null, "", "?" + params.toString());
    setSearchParams(newSearchParams);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const newRowsPerPage = parseInt(event.target.value, 10);
    const newSearchParams = {
      ...searchParams,
      rowsPerPage: newRowsPerPage,
      page: 0,
    };
    const params = toURLSearchParams(newSearchParams);
    window.history.replaceState(null, "", "?" + params.toString());
    setSearchParams(newSearchParams);
  };

  return (
    <Box>
      <HackTable overviews={hackList.hacks} />
      <TablePagination
        component="div"
        count={hackList.count}
        page={searchParams.page}
        onPageChange={handleChangePage}
        rowsPerPage={searchParams.rowsPerPage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        rowsPerPageOptions={[50, 100, 200]}
      />
    </Box>
  );
};

export default Hacks;
