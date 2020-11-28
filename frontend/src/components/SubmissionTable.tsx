import {
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow
} from "@material-ui/core";
import { green } from "@material-ui/core/colors";
import { DoneOutline } from "@material-ui/icons";
import "katex/dist/katex.min.css";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import { Link } from "react-router-dom";
import library_checker_client from "../api/library_checker_client";
import {
  LangListRequest,
  LangListResponse,
  SubmissionOverview
} from "../api/library_checker_pb";
import KatexRender from "./KatexRender";

interface OuterProps {
  overviews: SubmissionOverview[];
}
interface InnerProps {
  overviews: SubmissionOverview[];
  langListFetch: PromiseState<LangListResponse>;
}

const SubmissionTable: React.FC<InnerProps> = props => {
  const { overviews, langListFetch } = props;

  if (!langListFetch.fulfilled) {
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
  const idToName = langListFetch.value
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
          {overviews.map(row => (
            <TableRow key={row.getId()}>
              <TableCell>
                <Link to={`/submission/${row.getId()}`}>{row.getId()}</Link>
              </TableCell>
              <TableCell>
                <KatexRender text={row.getProblemTitle()} />
              </TableCell>
              <TableCell>{idToName[row.getLang()]}</TableCell>
              <TableCell>
                {row.getUserName() === "" ? "(Anonymous)" : row.getUserName()}
              </TableCell>
              <TableCell>
                {row.getStatus()}
                {row.getIsLatest() && row.getStatus() === "AC" && (
                  <DoneOutline style={{ color: green[500], height: "15px" }} />
                )}
              </TableCell>
              <TableCell>{Math.round(row.getTime() * 1000)} ms</TableCell>
              <TableCell>
                {row.getMemory() === -1 ? -1 : row.getMemory() / 1024 / 1024}{" "}
                MiB
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default connect<OuterProps, InnerProps>(props => ({
  langListFetch: {
    comparison: null,
    value: () => library_checker_client.langList(new LangListRequest())
  }
}))(SubmissionTable);
