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
import { useLangList } from "../api/client_wrapper";
import { SubmissionOverview } from "../api/library_checker";
import KatexTypography from "./katex/KatexTypography";
import { styled } from "@mui/system";
import { Timestamp } from "../api/google/protobuf/timestamp";

interface Props {
  overviews: SubmissionOverview[];
}

const CustomLink = styled(Link)(({ theme }) => ({
  color: theme.palette.primary.main,
  textDecoration: "none",
  textTransform: "none",
}));

const SubmissionTable: React.FC<Props> = (props) => {
  const { overviews } = props;

  const langListQuery = useLangList();

  if (langListQuery.isLoading || langListQuery.isError) {
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
  const idToName = langListQuery.data.langs.reduce<{ [name: string]: string }>(
    (dict, problem) => {
      dict[problem.id] = problem.name;
      return dict;
    },
    {}
  );

  return (
    <TableContainer>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Date</TableCell>
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
            <TableRow key={row.id}>
              <TableCell>
                <CustomLink to={`/submission/${row.id}`}>{row.id}</CustomLink>
              </TableCell>
              <TableCell>{row.submissionTime ? Timestamp.toDate(row.submissionTime).toLocaleString() : "-"}</TableCell>
              <TableCell>
                <CustomLink to={`/problem/${row.problemName}`}>
                  <KatexTypography>{row.problemTitle}</KatexTypography>
                </CustomLink>
              </TableCell>
              <TableCell>{idToName[row.lang]}</TableCell>
              <TableCell>
                {row.userName === "" ? (
                  "(Anonymous)"
                ) : (
                  <CustomLink to={`/user/${row.userName}`}>
                    {row.userName}
                  </CustomLink>
                )}
              </TableCell>
              <TableCell>
                {row.status}
                {row.isLatest && row.status === "AC" && (
                  <DoneOutline style={{ color: green[500], height: "15px" }} />
                )}
              </TableCell>
              <TableCell>{Math.round(row.time * 1000)} ms</TableCell>
              <TableCell>
                {row.memory === -1n
                  ? -1
                  : (Number(row.memory) / 1024 / 1024).toFixed(2)}{" "}
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
