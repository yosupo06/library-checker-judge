import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Button,
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
import React, { useContext, useState } from "react";
import { RouteComponentProps } from "react-router-dom";
import {
  RejudgeRequest,
  SubmissionInfoRequest,
} from "../api/library_checker_pb";
import SubmissionTable from "../components/SubmissionTable";
import Editor from "../components/Editor";
import { AuthContext } from "../contexts/AuthContext";
import library_checker_client, {
  authMetadata,
} from "../api/library_checker_client";
import { useQuery } from "react-query";

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

const OuterSubmissionInfo: React.FC<
  RouteComponentProps<{ submissionId: string }>
> = (props) => {
  const classes = useStyles();
  const auth = useContext(AuthContext);

  const submissionId = parseInt(props.match.params.submissionId);

  const [autoRefresh, setAutoRefresh] = useState(true);
  const submissionInfoQuery = useQuery(
    ["submissionInfo", submissionId],
    () =>
      library_checker_client.submissionInfo(
        new SubmissionInfoRequest().setId(submissionId),
        (auth ? authMetadata(auth.state) : null) ?? null
      ),
    {
      refetchInterval: autoRefresh ? 1000 : false,
      onSuccess: () => {
        const status = submissionInfoQuery.data?.getOverview()?.getStatus();
        if (
          status &&
          new Set(["AC", "WA", "RE", "TLE", "PE", "Fail", "CE", "IE"]).has(
            status
          )
        ) {
          setAutoRefresh(false);
        }
      },
    }
  );

  if (submissionInfoQuery.isLoading || submissionInfoQuery.isIdle) {
    return <h1>Loading</h1>;
  }
  if (submissionInfoQuery.isError) {
    return <h1>Error</h1>;
  }

  const handleRejudge = (e: React.FormEvent) => {
    e.preventDefault();
    library_checker_client
      .rejudge(
        new RejudgeRequest().setId(submissionId),
        (auth ? authMetadata(auth.state) : null) ?? null
      )
      .then(() => {
        console.log("Rejudge requested");
      });
  };

  const info = submissionInfoQuery.data;
  const compileError = new TextDecoder().decode(info.getCompileError_asU8());
  const overview = info.getOverview();
  const lang = overview ? overview.getLang() : undefined;

  return (
    <Box>
      <Box className={classes.overviewBox}>
        <Typography variant="h2" paragraph={true}>
          Submission Info #{overview?.getId()}
        </Typography>
        {info.getCanRejudge() && (
          <form onSubmit={(e) => handleRejudge(e)}>
            <Button color="primary" type="submit">
              Rejudge
            </Button>
          </form>
        )}
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

export default OuterSubmissionInfo;
