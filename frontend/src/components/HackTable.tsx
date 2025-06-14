import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import React from "react";
import { Link } from "react-router-dom";
import { HackOverview } from "../proto/library_checker";
import { styled } from "@mui/system";
import { Timestamp } from "../proto/google/protobuf/timestamp";

interface Props {
  overviews: HackOverview[];
}

const CustomLink = styled(Link)(({ theme }) => ({
  color: theme.palette.primary.main,
  textDecoration: "none",
  textTransform: "none",
}));

const HackTable: React.FC<Props> = (props) => {
  const { overviews } = props;

  const formatDate = (timestamp: Timestamp) => {
    const date = new Date(Number(timestamp.seconds) * 1000);
    return date.toLocaleString();
  };

  const formatTime = (time?: number) => {
    if (time === undefined) return "-";
    return `${time.toFixed(3)}s`;
  };

  const formatMemory = (memory?: bigint) => {
    if (memory === undefined) return "-";
    const memoryKB = Number(memory) / 1024;
    return `${memoryKB.toFixed(0)} KB`;
  };

  return (
    <TableContainer>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Hack ID</TableCell>
            <TableCell>Submission ID</TableCell>
            <TableCell>User</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Time</TableCell>
            <TableCell>Memory</TableCell>
            <TableCell>Hack Time</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {overviews.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7} align="center">
                No hacks found
              </TableCell>
            </TableRow>
          ) : (
            overviews.map((overview) => (
              <TableRow key={overview.id}>
                <TableCell>
                  <CustomLink to={`/hack/${overview.id}`}>
                    {overview.id}
                  </CustomLink>
                </TableCell>
                <TableCell>
                  <CustomLink to={`/submission/${overview.submissionId}`}>
                    {overview.submissionId}
                  </CustomLink>
                </TableCell>
                <TableCell>
                  {overview.userName ? (
                    <CustomLink to={`/user/${overview.userName}`}>
                      {overview.userName}
                    </CustomLink>
                  ) : (
                    "Anonymous"
                  )}
                </TableCell>
                <TableCell>{overview.status}</TableCell>
                <TableCell>{formatTime(overview.time)}</TableCell>
                <TableCell>{formatMemory(overview.memory)}</TableCell>
                <TableCell>
                  {overview.hackTime ? formatDate(overview.hackTime) : "-"}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default HackTable;
