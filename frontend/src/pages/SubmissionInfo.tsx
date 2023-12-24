import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import Accordion from "@mui/material/Accordion";
import AccordionDetails from "@mui/material/AccordionDetails";
import AccordionSummary from "@mui/material/AccordionSummary";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { Autorenew, ExpandMore } from "@mui/icons-material";
import React from "react";
import { useParams } from "react-router-dom";
import SubmissionTable from "../components/SubmissionTable";
import SourceEditor from "../components/SourceEditor";
import {
  useRejudgeMutation,
  useSubmissionInfo,
  useUserInfo,
} from "../api/client_wrapper";
import CircularProgress from "@mui/material/CircularProgress";
import Link from "@mui/material/Link";
import { LibraryBooks } from "@mui/icons-material";
import { Alert, Collapse, Container, Divider, IconButton } from "@mui/material";
import {
  SubmissionCaseResult,
  SubmissionInfoResponse,
} from "../proto/library_checker";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import { RpcError } from "@protobuf-ts/runtime-rpc";

const SubmissionInfo: React.FC = () => {
  const { submissionId } = useParams<"submissionId">();
  if (!submissionId) {
    throw new Error(`submissionId is not defined`);
  }
  const submissionIdInt = parseInt(submissionId);

  const submissionInfoQuery = useSubmissionInfo(submissionIdInt, {
    refetchInterval: 1000,
  });

  if (submissionInfoQuery.isLoading) {
    return (
      <Container>
        <CircularProgress />
      </Container>
    );
  }

  if (submissionInfoQuery.isError) {
    return (
      <Container>
        <Alert severity="error">
          {(submissionInfoQuery.error as RpcError).toString()}
        </Alert>
      </Container>
    );
  }
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Submission #{submissionId}
      </Typography>
      <SubmissionInfoBody info={submissionInfoQuery.data} />
    </Container>
  );
};

export default SubmissionInfo;

const SubmissionInfoBody: React.FC<{
  info: SubmissionInfoResponse;
}> = (props) => {
  const { info } = props;
  const compileError = new TextDecoder().decode(info.compileError);
  const overview = info.overview;
  const lang = overview ? overview.lang : undefined;

  return (
    <Box>
      <Rejudge info={info} />
      <Divider
        sx={{
          marginTop: 3,
          marginBottom: 3,
        }}
      />
      <Overview info={info} />
      <Divider
        sx={{
          marginTop: 3,
          marginBottom: 3,
        }}
      />
      {compileError && (
        <Paper>
          <Accordion>
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography>Compile Error</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <pre>{compileError}</pre>
            </AccordionDetails>
          </Accordion>
        </Paper>
      )}
      <CaseResults info={info} />
      <Divider
        sx={{
          marginTop: 3,
          marginBottom: 3,
        }}
      />
      <Paper>
        <SourceEditor
          value={info.source}
          language={lang}
          readOnly={true}
          autoHeight={true}
        />
      </Paper>
    </Box>
  );
};

const Rejudge: React.FC<{ info: SubmissionInfoResponse }> = (props) => {
  const { info } = props;
  const overview = info.overview;
  if (!overview) {
    return <Box>Loading error</Box>;
  }

  const mutation = useRejudgeMutation();

  const handleRejudge = (e: React.FormEvent) => {
    e.preventDefault();

    mutation.mutate({
      id: overview.id,
    });
  };

  return (
    <Box>
      {info.canRejudge && (
        <Button
          variant="outlined"
          startIcon={<Autorenew />}
          onClick={handleRejudge}
        >
          Rejudge
        </Button>
      )}
      {mutation.isSuccess && (
        <Alert severity="success">Sent rejudge request</Alert>
      )}
      {mutation.isError && (
        <Alert severity="error">{(mutation.error as RpcError).message}</Alert>
      )}
    </Box>
  );
};

const Overview: React.FC<{ info: SubmissionInfoResponse }> = (props) => {
  const { info } = props;
  const overview = info.overview;
  if (!overview) {
    return <Box>Loading error</Box>;
  }

  return (
    <Box>
      <SubmissionTable overviews={[overview]} />
      <Box
        sx={{
          marginTop: 1,
        }}
      >
        {overview.userName && <LibraryButton name={overview.userName} />}
      </Box>
    </Box>
  );
};

const LibraryButton: React.FC<{ name: string }> = (props) => {
  const userInfoQuery = useUserInfo(props.name, {});

  if (userInfoQuery.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (userInfoQuery.isError) {
    return <Box>Failed to load user</Box>;
  }

  const userInfo = userInfoQuery.data;
  const user = userInfo.user;

  if (!user) {
    return <Box>Failed to load user</Box>;
  }

  const libraryURL = user.libraryUrl;

  if (!libraryURL) {
    return <Box></Box>;
  }
  return (
    <Button variant="outlined" startIcon={<LibraryBooks />}>
      <Link href={libraryURL}> {libraryURL}</Link>
    </Button>
  );
};

const toStringAsUTF8 = (data: Uint8Array) => {
  return new TextDecoder("utf-8", { fatal: false }).decode(data);
};

const CaseResults: React.FC<{ info: SubmissionInfoResponse }> = (props) => {
  const { info } = props;

  return (
    <Box>
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography>Case Results</Typography>
        </AccordionSummary>
        <AccordionDetails>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow key={"details"}>
                  <TableCell />
                  <TableCell>Name</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Time</TableCell>
                  <TableCell>Memory</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {info.caseResults.map((row) => (
                  <CaseResultRow key={row.case} row={row} />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </AccordionDetails>
      </Accordion>
    </Box>
  );
};
const CaseResultRow: React.FC<{ row: SubmissionCaseResult }> = (props) => {
  const { row } = props;
  const [open, setOpen] = React.useState(false);

  return (
    <>
      <TableRow>
        <TableCell>
          <IconButton
            aria-label="expand row"
            size="small"
            onClick={() => setOpen(!open)}
          >
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell>{row.case}</TableCell>
        <TableCell>{row.status}</TableCell>
        <TableCell>{Math.round(row.time * 1000)} ms</TableCell>
        <TableCell>
          {row.memory === -1n
            ? -1
            : (Number(row.memory) / 1024 / 1024).toFixed(2)}{" "}
          Mib
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              <Typography variant="h6" gutterBottom>
                Stderr
              </Typography>
              <pre>{toStringAsUTF8(row.stderr)}</pre>
              <Typography variant="h6" gutterBottom>
                Checker Output
              </Typography>
              <pre>{toStringAsUTF8(row.checkerOut)}</pre>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
};
