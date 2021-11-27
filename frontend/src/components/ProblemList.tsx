import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import React from "react";
import { Link } from "react-router-dom";
import KatexRender from "./KatexRender";
import { lightGreen, cyan } from "@mui/material/colors";
interface Props {
  problems: {
    name: string;
    title: string;
    status?: "ac";
  }[];
  solvedStatus: {
    [problem: string]: "latest_ac" | "ac" | "unknown";
  };
}

const ProblemList: React.FC<Props> = (props) => {
  const { problems, solvedStatus } = props;

  const bgColorMap = {
    latest_ac: lightGreen.A200,
    ac: cyan.A400,
    unknown: undefined,
  };

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableBody>
          {problems.map((problem) => (
            <TableRow key={problem.name}>
              <TableCell
                sx={{
                  bgcolor: bgColorMap[solvedStatus[problem.name]],
                }}
              >
                <Link to={`/problem/${problem.name}`}>
                  <KatexRender text={problem.title} />
                </Link>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default ProblemList;
