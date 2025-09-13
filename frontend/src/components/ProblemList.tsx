import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import React from "react";
import type { components as OpenApi } from "../openapi/types";
import { Link } from "react-router-dom";
import { lightGreen, cyan } from "@mui/material/colors";
import KatexTypography from "./katex/KatexTypography";
import { styled } from "@mui/material/styles";
interface Props {
  problems: OpenApi["schemas"]["Problem"][];
  solvedStatus?: {
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
                  bgcolor: solvedStatus
                    ? bgColorMap[solvedStatus[problem.name]]
                    : undefined,
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
