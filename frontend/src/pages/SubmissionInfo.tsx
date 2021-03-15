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
  Typography,
} from "@material-ui/core";
import { ExpandMore } from "@material-ui/icons";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import { RouteComponentProps } from "react-router-dom";
import library_checker_client from "../api/library_checker_client";
import {
  SubmissionInfoRequest,
  SubmissionInfoResponse,
} from "../api/library_checker_pb";
import SubmissionTable from "../components/SubmissionTable";
import Editor from "../components/Editor";

interface Props {
  submissionInfoFetch: PromiseState<SubmissionInfoResponse>;
  fixSubmissionInfo: (value: SubmissionInfoResponse) => void;
}

const useStyles = makeStyles((theme) => ({
  overviewBox: {
    marginBottom: theme.spacing(0.2),
  },
  compileErrorText: {
    whiteSpace: "pre",
    fontSize: "11px",
    width: "100%",
  },
}));

const SubmissionInfo: React.FC<Props> = (props) => {
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
  const overview = info.getOverview();
  const status = overview ? overview.getStatus() : undefined;
  const lang = overview ? overview.getLang() : undefined;

  if (
    status &&
    new Set(["AC", "WA", "RE", "TLE", "PE", "Fail", "CE", "IE"]).has(status)
  ) {
    props.fixSubmissionInfo(info);
  }

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
                <pre className={classes.compileErrorText}>{compileError}</pre>
              </AccordionDetails>
            </Accordion>
          </Paper>
        )}
        {info.getCaseResultsList().length !== 0 && (
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
                      {info.getCaseResultsList().map((row) => (
                        <TableRow key={row.getCase()}>
                          <TableCell>{row.getCase()}</TableCell>
                          <TableCell>{row.getStatus()}</TableCell>
                          <TableCell>
                            {Math.round(row.getTime() * 1000)} ms
                          </TableCell>
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
              </AccordionDetails>
            </Accordion>
          </Paper>
        )}
      </Box>
      <Paper>
        <Editor
          value={info.getSource()}
          language={lang}
          readOnly={true}
          autoHeight={true}
        />
      </Paper>
    </Box>
  );
};

export default connect<RouteComponentProps<{ submissionId: string }>, Props>(
  (props) => ({
    submissionInfoFetch: {
      comparison: null,
      refreshInterval: 2000,
      value: () =>
        library_checker_client.submissionInfo(
          new SubmissionInfoRequest().setId(
            parseInt(props.match.params.submissionId)
          ),
          {}
        ),
    },
    fixSubmissionInfo: (value: SubmissionInfoResponse) => ({
      submissionInfoFetch: {
        refreshing: true,
        value: value,
      },
    }),
  })
)(SubmissionInfo);
