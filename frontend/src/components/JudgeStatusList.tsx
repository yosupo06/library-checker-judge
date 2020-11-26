import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow
} from "@material-ui/core";
import React from "react";

const JudgeStatusList = () => {
  const judge_status = [
    {
      name: "AC",
      text: "Accepted (Green: latest testcase)"
    },
    {
      name: "WA",
      text: "Wrong Answer"
    },
    {
      name: "RE",
      text: "Runtime Error"
    },
    {
      name: "TLE",
      text: "Time Limit Exceeded"
    },
    {
      name: "PE",
      text: "Presentation Error"
    },
    {
      name: "Fail",
      text: "An author's solution is wrong"
    },
    {
      name: "CE",
      text: "Compile Error"
    },
    {
      name: "WJ",
      text: "Waiting Judge"
    }
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
          {judge_status.map(row => (
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
