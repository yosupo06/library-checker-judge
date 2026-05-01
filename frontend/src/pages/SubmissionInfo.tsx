import Box from "@mui/material/Box";
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
import Input from "@mui/icons-material/Input";
import Autorenew from "@mui/icons-material/Autorenew";
import ExpandMore from "@mui/icons-material/ExpandMore";
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
import { LibraryBooks } from "@mui/icons-material";
import { Alert, Collapse, Divider, IconButton } from "@mui/material";
import {
  SubmissionCaseResult,
  SubmissionInfoResponse,
} from "../proto/library_checker";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import MainContainer from "../components/MainContainer";
import { ExternalLinkButton, LinkButton } from "../components/LinkButton";
import NotFound from "./NotFound";

const SubmissionInfo: React.FC = () => {
  const { submissionId } = useParams<"submissionId">();
  if (!submissionId) {
    throw new Error(`submissionId is not defined`);
  }
  const intID = parseInt(submissionId);
  if (Number.isNaN(intID)) {
    return <NotFound />;
  }

  return (
    <MainContainer title={`Submission #${submissionId}`}>
      <SubmissionInfoBody id={intID} />
    </MainContainer>
  );
};

export default SubmissionInfo;

const SubmissionInfoBody: React.FC<{
  id: number;
}> = (props) => {
  const { id } = props;
  const submissionInfoQuery = useSubmissionInfo(id, {
    refetchInterval: 1000,
  });

  if (submissionInfoQuery.isPending) {
    return <CircularProgress />;
  }
  if (submissionInfoQuery.isError) {
    return <Alert severity="error">{submissionInfoQuery.error.message}</Alert>;
  }

  const info = submissionInfoQuery.data;
  const overview = info.overview;
  if (!overview) {
    return <Alert severity="error">Submission overview is not found</Alert>;
  }

  const compileError = new TextDecoder().decode(info.compileError);
  const lang = overview ? overview.lang : undefined;

  return (
    <>
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
        <SourceEditor value={info.source} language={lang} readOnly={true} />
      </Paper>
    </>
  );
};

const Overview: React.FC<{ info: SubmissionInfoResponse }> = (props) => {
  const { info } = props;
  const overview = info.overview!;

  return (
    <>
      <SubmissionTable overviews={[overview]} />
      <Box
        sx={{
          marginTop: 1,
        }}
      >
        <UsefulLinks info={info} />
      </Box>
    </>
  );
};

const UsefulLinks: React.FC<{
  info: SubmissionInfoResponse;
}> = (props) => {
  const { info } = props;
  const overview = info.overview!;

  const mutation = useRejudgeMutation();
  const handleRejudge = (e: React.FormEvent) => {
    e.preventDefault();

    mutation.mutate({
      id: overview.id,
    });
  };

  return (
    <>
      {mutation.isSuccess && (
        <Alert severity="success">Sent rejudge request</Alert>
      )}
      {mutation.isError && (
        <Alert severity="error">{mutation.error.message}</Alert>
      )}
      <LinkButton
        variant="outlined"
        startIcon={<Input />}
        to={`/hack/?id=${overview.id}`}
      >
        Hack
      </LinkButton>
      {info.canRejudge && (
        <LinkButton
          variant="outlined"
          startIcon={<Autorenew />}
          onClick={handleRejudge}
        >
          Rejudge
        </LinkButton>
      )}
      {overview.userName && <LibraryButton name={overview.userName} />}
    </>
  );
};

const LibraryButton: React.FC<{ name: string }> = (props) => {
  const userInfoQuery = useUserInfo(props.name, {});

  if (userInfoQuery.isPending) {
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
    return <></>;
  }
  return (
    <ExternalLinkButton startIcon={<LibraryBooks />} href={libraryURL}>
      {libraryURL}
    </ExternalLinkButton>
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
          <Typography>Test cases</Typography>
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
