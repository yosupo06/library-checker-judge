import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  makeStyles,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography
} from "@material-ui/core";
import { ExpandMore } from "@material-ui/icons";
import React from "react";
import AceEditor from "react-ace";
import { connect, PromiseState } from "react-refetch";
import { RouteComponentProps } from "react-router-dom";
import library_checker_client from "../api/library_checker_client";
import {
  SubmissionInfoRequest,
  SubmissionInfoResponse
} from "../api/library_checker_pb";
import SubmissionTable from "../components/SubmissionTable";

import "ace-builds/src-min-noconflict/mode-csharp";
import "ace-builds/src-min-noconflict/mode-c_cpp";
import "ace-builds/src-min-noconflict/mode-d";
import "ace-builds/src-min-noconflict/mode-golang";
import "ace-builds/src-min-noconflict/mode-haskell";
import "ace-builds/src-min-noconflict/mode-java";
import "ace-builds/src-min-noconflict/mode-lisp";
import "ace-builds/src-min-noconflict/mode-python";
import "ace-builds/src-min-noconflict/mode-rust";
import "ace-builds/src-min-noconflict/mode-text";

interface Props {
  submissionInfoFetch: PromiseState<SubmissionInfoResponse>;
  fixSubmissionInfo: (value: SubmissionInfoResponse) => void;
}

const useStyles = makeStyles(theme => ({
  overviewBox: {
    marginBottom: theme.spacing(0.2)
  },
  compileErrorText: {
    whiteSpace: "pre",
    fontSize: "11px",
    fontFamily: '"Courier New", Consolas, monospace',
    background: theme.palette.grey[50],
    width: "100%"
  }
}));

const SubmissionInfo: React.FC<Props> = props => {
  const { submissionInfoFetch } = props;
  const classes = useStyles();

  if (submissionInfoFetch.pending) {
    return <h1>Loading</h1>;
  }
  if (submissionInfoFetch.rejected) {
    return <h1>Error</h1>;
  }
  const info = submissionInfoFetch.value;
  const compileError = new TextDecoder().decode(info.getCompileError_asU8());
  const overview = info.getOverview()!;
  const status = overview.getStatus();
  const lang = overview.getLang()!;

  if (
    new Set(["AC", "WA", "RE", "TLE", "PE", "Fail", "CE", "IE"]).has(status)
  ) {
    props.fixSubmissionInfo(info);
  }

  const aceMode = (() => {
    if (lang.startsWith("cpp")) {
      return "c_cpp";
    }
    if (lang.startsWith("java")) {
      return "java";
    }
    if (lang.startsWith("py")) {
      return "python";
    }
    if (lang.startsWith("rust")) {
      return "rust";
    }
    if (lang.startsWith("d")) {
      return "d";
    }
    if (lang.startsWith("haskell")) {
      return "haskell";
    }
    if (lang.startsWith("csharp")) {
      return "csharp";
    }
    if (lang.startsWith("go")) {
      return "golang";
    }
    if (lang.startsWith("lisp")) {
      return "lisp";
    }
    return "text";
  })();

  return (
    <Box>
      <Box className={classes.overviewBox}>
        <Typography variant="h2" paragraph={true}>
          Submission Info #{overview?.getId()}
        </Typography>
        {overview && (
          <Paper>
            <SubmissionTable overviews={[overview]} />
          </Paper>
        )}
        {compileError && (
          <Paper>
            <Accordion>
              <AccordionSummary expandIcon={<ExpandMore />}>
                <Typography>Compile Error</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Typography className={classes.compileErrorText}>
                  {compileError}
                </Typography>
              </AccordionDetails>
            </Accordion>
          </Paper>
        )}
        {info.getCaseResultsList().length && (
          <Paper>
            <Accordion defaultExpanded>
              <AccordionSummary expandIcon={<ExpandMore />}>
                <Typography>Case Result</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <TableContainer>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Name</TableCell>
                        <TableCell>Status</TableCell>
                        <TableCell>Time</TableCell>
                        <TableCell>Memory</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {info.getCaseResultsList().map(row => (
                        <TableRow key={row.getCase()}>
                          <TableCell>{row.getCase()}</TableCell>
                          <TableCell>{row.getStatus()}</TableCell>
                          <TableCell>
                            {Math.round(row.getTime() * 1000)} ms
                          </TableCell>
                          <TableCell>
                            {row.getMemory() === -1
                              ? -1
                              : row.getMemory() / 1024 / 1024}{" "}
                            Mib
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </AccordionDetails>
            </Accordion>
          </Paper>
        )}
        {compileError && (
          <Paper>
            <Accordion>
              <AccordionSummary expandIcon={<ExpandMore />}>
                <Typography>Compile Error</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Typography className={classes.compileErrorText}>
                  {compileError}
                </Typography>
              </AccordionDetails>
            </Accordion>
          </Paper>
        )}
      </Box>
      <AceEditor
        mode={aceMode}
        value={info.getSource()}
        maxLines={Infinity}
        readOnly={true}
        width="100%"
        showPrintMargin={false}
      />
    </Box>
  );
};

export default connect<RouteComponentProps<{ submissionId: string }>, Props>(
  props => ({
    submissionInfoFetch: {
      comparison: null,
      refreshInterval: 2000,
      value: () =>
        library_checker_client.submissionInfo(
          new SubmissionInfoRequest().setId(
            parseInt(props.match.params.submissionId)
          )
        )
    },
    fixSubmissionInfo: (value: SubmissionInfoResponse) => ({
      submissionInfoFetch: {
        refreshing: true,
        value: value
      }
    })
  })
)(SubmissionInfo);
