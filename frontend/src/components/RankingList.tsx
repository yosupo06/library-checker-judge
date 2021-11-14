import { Box, Typography } from "@material-ui/core";
import { GridColDef, DataGrid } from "@material-ui/data-grid";
import React from "react";
import { useQuery } from "react-query";
import library_checker_client from "../api/library_checker_client";
import { RankingRequest } from "../api/library_checker_pb";

const RankingList: React.FC = () => {
  const rankingQuery = useQuery("ranking", () =>
    library_checker_client.ranking(new RankingRequest(), {})
  );

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
