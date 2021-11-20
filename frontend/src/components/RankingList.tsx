import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { DataGrid, GridColDef } from "@mui/x-data-grid";
import React from "react";
import { useRanking } from "../api/library_checker_client";

const RankingList: React.FC = () => {
  const rankingQuery = useRanking();

  if (rankingQuery.isLoading || rankingQuery.isIdle) {
    return (
      <Box>
        <Typography>Loading...</Typography>
      </Box>
    );
  }
  if (rankingQuery.isError) {
    return (
      <Box>
        <Typography>Error: {rankingQuery.error}</Typography>
      </Box>
    );
  }

  const columns: GridColDef[] = [
    { field: "name", headerName: "ID", width: 130 },
    { field: "count", headerName: "AC Count" },
  ];
  const rows = rankingQuery.data.getStatisticsList().map((e, index) => {
    return {
      id: index,
      name: e.getName(),
      count: e.getCount(),
    };
  });
  return (
    <Box style={{ height: 2000 }}>
      <DataGrid rows={rows} columns={columns} />
    </Box>
  );
};

export default RankingList;
