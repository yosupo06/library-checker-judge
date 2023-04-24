import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import React from "react";
import { Problem } from "../api/library_checker";
import { Link } from "react-router-dom";
import { lightGreen, cyan } from "@mui/material/colors";
import KatexTypography from "./katex/KatexTypography";
import { styled } from "@mui/system";
interface Props {
  problems: Problem[];
  solvedStatus: {
    [problem: string]: "latest_ac" | "ac" | "unknown";
  };
}

const NavbarLink = styled(Link)({
  color: "inherit",
  textDecoration: "none",
  textTransform: "none",
});

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
                <NavbarLink to={`/problem/${problem.name}`}>
                  <KatexTypography>{problem.title}</KatexTypography>
                </NavbarLink>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default ProblemList;
