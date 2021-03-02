import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@material-ui/core";
import React from "react";
import { Link } from "react-router-dom";
import KatexRender from "./KatexRender";

interface Props {
  problems: {
    name: string;
    title: string;
    status?: "ac";
  }[];
}

const ProblemList: React.FC<Props> = (props) => {
  const { problems } = props;

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
              <TableCell>
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
