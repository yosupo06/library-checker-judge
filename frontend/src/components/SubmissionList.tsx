import { Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core';
import 'katex/dist/katex.min.css';
import React from 'react';
import { SubmissionOverview } from "../api/library_checker_pb";
import { Link, RouteComponentProps, withRouter } from 'react-router-dom';
import KatexRender from './KatexRender';
import { DoneOutline } from '@material-ui/icons';
import { green } from '@material-ui/core/colors';

interface Props {
  submissionOverviews: SubmissionOverview[];
}

const SubmissionList: React.FC<RouteComponentProps & Props> = (props) => {
  const { history } = props
  const { submissionOverviews } = props;

  return (
    <TableContainer component={Paper}>
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
          {submissionOverviews.map((row) => (
            <TableRow key={row.getId()}>
              <TableCell><Link to={`/submission/${row.getId()}`}>{row.getId()}</Link></TableCell>
              <TableCell><KatexRender text={row.getProblemTitle()} /></TableCell>
              <TableCell>{row.getLang()}</TableCell>
              <TableCell>{row.getUserName() === "" ? "(Anonymous)" : row.getUserName()}</TableCell>
              <TableCell>{row.getIsLatest() && <DoneOutline style={{ color: green[500] }} />}{row.getStatus()}</TableCell>
              <TableCell>{Math.round(row.getTime() * 1000)} ms</TableCell>
              <TableCell>{row.getMemory() === -1 ? -1 : row.getMemory() / 1024 / 1024} Mib</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

export default withRouter(SubmissionList);
