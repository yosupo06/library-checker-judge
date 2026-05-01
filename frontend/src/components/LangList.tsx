import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import TableHead from "@mui/material/TableHead";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import React from "react";
import { useLangList } from "../api/client_wrapper";

const LangList: React.FC = () => {
  const langListQuery = useLangList();

  if (langListQuery.isPending) {
    return (
      <Box>
        <Typography>Loading...</Typography>
      </Box>
    );
  }
  if (langListQuery.isError) {
    return (
      <Box>
        <Typography>Error: {langListQuery.error.message}</Typography>
      </Box>
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
          {langList.langs.map((row) => (
            <TableRow key={row.name}>
              <TableCell>{row.name}</TableCell>
              <TableCell>{row.version}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default LangList;
