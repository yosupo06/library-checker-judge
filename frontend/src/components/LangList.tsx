import {
  Container,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography
} from "@material-ui/core";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import library_checker_client from "../api/library_checker_client";
import { LangListRequest, LangListResponse } from "../api/library_checker_pb";

interface Props {
  langListFetch: PromiseState<LangListResponse>;
}

const LangList: React.FC<Props> = props => {
  const { langListFetch } = props;

  if (langListFetch.pending) {
    return (
      <Container>
        <Typography>Loading...</Typography>
      </Container>
    );
  }
  if (langListFetch.rejected) {
    return (
      <Container>
        <Typography>Error: {langListFetch.reason}</Typography>
      </Container>
    );
  }
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>Version</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {langListFetch.value.getLangsList().map(row => (
            <TableRow key={row.getName()}>
              <TableCell>{row.getName()}</TableCell>
              <TableCell>{row.getVersion()}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default connect<{}, Props>(() => ({
  langListFetch: {
    comparison: null,
    value: library_checker_client.langList(new LangListRequest())
  }
}))(LangList);
