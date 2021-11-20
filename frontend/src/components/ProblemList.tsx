import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { makeStyles, Theme } from "@material-ui/core";
import React from "react";
import { Link } from "react-router-dom";
import KatexRender from "./KatexRender";
import lightGreen from "@material-ui/core/colors/lightGreen";
import cyan from "@material-ui/core/colors/cyan";

const useStyles = makeStyles((theme: Theme) => ({
  default: {},
  latest_ac: {
    backgroundColor: lightGreen.A200,
  },
  ac: {
    backgroundColor: cyan.A400,
  },
}));
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
  const classes = useStyles(props);
  const { problems, solvedStatus } = props;

  const classNameMap = {
    latest_ac: classes.latest_ac,
    ac: classes.ac,
    unknown: classes.default,
  };

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {problems.map((problem) => (
            <TableRow key={problem.name}>
              <TableCell className={classNameMap[solvedStatus[problem.name]]}>
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
