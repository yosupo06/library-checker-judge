import { Box, Container, Typography } from "@material-ui/core";
import { GridColDef, DataGrid } from "@material-ui/data-grid";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import library_checker_client from "../api/library_checker_client";
import { RankingRequest, RankingResponse } from "../api/library_checker_pb";

interface OuterProps {}
interface InnerProps {
  rankingFetch: PromiseState<RankingResponse>;
}

const RankingList: React.FC<InnerProps> = (props) => {
  const { rankingFetch } = props;

  if (rankingFetch.pending) {
    return (
      <Container>
        <Typography>Loading...</Typography>
      </Container>
    );
  }
  if (rankingFetch.rejected) {
    return (
      <Container>
        <Typography>Error: {rankingFetch.reason}</Typography>
      </Container>
    );
  }

  const columns: GridColDef[] = [
    { field: "name", headerName: "ID", width: 130 },
    { field: "count", headerName: "AC Count" },
  ];
  const rows = rankingFetch.value.getStatisticsList().map((e, index) => {
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

export default connect<OuterProps, InnerProps>(() => ({
  rankingFetch: {
    comparison: null,
    value: () => library_checker_client.ranking(new RankingRequest()),
  },
}))(RankingList);
