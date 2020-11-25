import { Container, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography } from '@material-ui/core';
import React from 'react';
import { connect, PromiseState } from 'react-refetch';
import library_checker_client from '../api/library_checker_client';
import { RankingRequest, RankingResponse } from "../api/library_checker_pb";

interface Props {
    rankingFetch: PromiseState<RankingResponse>;
}

const RankingList: React.FC<Props> = (props) => {
  const { rankingFetch } = props;

  if (rankingFetch.pending) {
    return (
      <Container>
        <Typography>
                    Loading...
        </Typography>
      </Container>
    );
  }
  if (rankingFetch.rejected) {
    return (
      <Container>
        <Typography>
                    Error: {rankingFetch.reason}
        </Typography>
      </Container>
    )
  }
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>AC Count</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rankingFetch.value.getStatisticsList().map((row) => (
            <TableRow key={row.getName()}>
              <TableCell>{row.getName()}</TableCell>
              <TableCell>{row.getCount()}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

export default connect<{}, Props>(() => ({
  rankingFetch: {
    comparison: null,
    value: library_checker_client.ranking(new RankingRequest())
  }
}))(RankingList);
