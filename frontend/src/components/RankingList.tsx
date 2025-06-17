import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import React, { useState } from "react";
import { useRanking } from "../api/client_wrapper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Pagination from "@mui/material/Pagination";
import Stack from "@mui/material/Stack";

const RankingList: React.FC = () => {
  const [page, setPage] = useState(1);
  const limit = 50; // Users per page
  const skip = (page - 1) * limit;

  const rankingQuery = useRanking(skip, limit);

  if (rankingQuery.isPending) {
    return (
      <Box>
        <Typography>Loading...</Typography>
      </Box>
    );
  }
  if (rankingQuery.isError) {
    return (
      <Box>
        <Typography>Error: {rankingQuery.error.message}</Typography>
      </Box>
    );
  }

  const totalCount = rankingQuery.data.count;
  const totalPages = Math.ceil(totalCount / limit);

  const rows = rankingQuery.data.statistics.map((e, index) => {
    return {
      id: index,
      name: e.name,
      count: e.count,
      ranking: skip + index + 1, // Calculate actual ranking position
    };
  });

  const handlePageChange = (
    event: React.ChangeEvent<unknown>,
    value: number,
  ) => {
    setPage(value);
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Total Users: {totalCount}
      </Typography>

      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Ranking</TableCell>
              <TableCell>ID</TableCell>
              <TableCell>AC Count</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell>{row.ranking}</TableCell>
                <TableCell>{row.name}</TableCell>
                <TableCell>{row.count}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Stack spacing={2} alignItems="center" sx={{ marginTop: 2 }}>
        <Pagination
          count={totalPages}
          page={page}
          onChange={handlePageChange}
          color="primary"
          size="large"
          showFirstButton
          showLastButton
        />
      </Stack>
    </Box>
  );
};

export default RankingList;
