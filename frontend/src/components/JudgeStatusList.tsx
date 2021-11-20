import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import TableHead from "@mui/material/TableHead";
import React from "react";

const JudgeStatusList = (): JSX.Element => {
  const judge_status = [
    {
      name: "AC",
      text: "Accepted (Green check: in the latest testcase)",
    },
    {
      name: "WA",
      text: "Wrong Answer",
    },
    {
      name: "RE",
      text: "Runtime Error",
    },
    {
      name: "TLE",
      text: "Time Limit Exceeded",
    },
    {
      name: "PE",
      text: "Presentation Error",
    },
    {
      name: "Fail",
      text: "An author's solution is wrong",
    },
    {
      name: "CE",
      text: "Compile Error",
    },
    {
      name: "WJ",
      text: "Waiting Judge",
    },
    {
      name: "IE",
      text: "Judge Server is broken 😢",
    },
  ];

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Status</TableCell>
            <TableCell>Info</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {judge_status.map((row) => (
            <TableRow key={row.name}>
              <TableCell>{row.name}</TableCell>
              <TableCell>{row.text}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default JudgeStatusList;
