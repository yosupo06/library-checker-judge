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
import { useLangList } from "../api/library_checker_client";

const LangList: React.FC = () => {
  const langListQuery = useLangList();

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
  const langList = langListQuery.data;
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
          {langList.getLangsList().map((row) => (
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
