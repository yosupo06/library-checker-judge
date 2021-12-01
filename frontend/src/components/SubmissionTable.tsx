import CircularProgress from "@mui/material/CircularProgress";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { green } from "@mui/material/colors";
import { DoneOutline } from "@mui/icons-material";
import "katex/dist/katex.min.css";
import React from "react";
import { Link } from "react-router-dom";
import { useLangList } from "../api/library_checker_client";
import { SubmissionOverview } from "../api/library_checker_pb";
import KatexTypography from "./katex/KatexTypography";

interface Props {
  overviews: SubmissionOverview[];
}

const SubmissionTable: React.FC<Props> = (props) => {
  const { overviews } = props;

  const langListQuery = useLangList();

  if (
    langListQuery.isLoading ||
    langListQuery.isIdle ||
    langListQuery.isError
  ) {
    return (
      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>Problem</TableCell>
              <TableCell>Lang</TableCell>
              <TableCell>User</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Time</TableCell>
              <TableCell>Memory</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <CircularProgress />
          </TableBody>
        </Table>
      </TableContainer>
    );
  }
  const idToName = langListQuery.data
    .getLangsList()
    .reduce<{ [name: string]: string }>((dict, problem) => {
      dict[problem.getId()] = problem.getName();
      return dict;
    }, {});

  return (
    <TableContainer>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Problem</TableCell>
            <TableCell>Lang</TableCell>
            <TableCell>User</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Time</TableCell>
            <TableCell>Memory</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {overviews.map((row) => (
            <TableRow key={row.getId()}>
              <TableCell>
                <Link to={`/submission/${row.getId()}`}>{row.getId()}</Link>
              </TableCell>
              <TableCell>
                <Link to={`/problem/${row.getProblemName()}`}>
                  <KatexTypography>{row.getProblemTitle()}</KatexTypography>
                </Link>
              </TableCell>
              <TableCell>{idToName[row.getLang()]}</TableCell>
              <TableCell>
                {row.getUserName() === "" ? (
                  "(Anonymous)"
                ) : (
                  <Link to={`/user/${row.getUserName()}`}>
                    {row.getUserName()}
                  </Link>
                )}
              </TableCell>
              <TableCell>
                {row.getStatus()}
                {row.getIsLatest() && row.getStatus() === "AC" && (
                  <DoneOutline style={{ color: green[500], height: "15px" }} />
                )}
              </TableCell>
              <TableCell>{Math.round(row.getTime() * 1000)} ms</TableCell>
              <TableCell>
                {row.getMemory() === -1
                  ? -1
                  : (row.getMemory() / 1024 / 1024).toFixed(2)}{" "}
                Mib
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default SubmissionTable;
