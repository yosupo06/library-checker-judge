import {
  Container,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@material-ui/core";
import React from "react";
import { useQuery } from "react-query";
import library_checker_client from "../api/library_checker_client";
import { LangListRequest } from "../api/library_checker_pb";

const LangList: React.FC = () => {
  const langListQuery = useQuery("langList", () =>
    library_checker_client.langList(new LangListRequest(), {})
  );

  if (langListQuery.isLoading || langListQuery.isIdle) {
    return (
      <Container>
        <Typography>Loading...</Typography>
      </Container>
    );
  }
  if (langListQuery.isError) {
    return (
      <Container>
        <Typography>Error: {langListQuery.error}</Typography>
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
          {langListQuery.data.getLangsList().map((row) => (
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

export default LangList;
