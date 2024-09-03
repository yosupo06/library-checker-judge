import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import React from "react";
import { useParams } from "react-router-dom";
import { useHackInfo } from "../api/client_wrapper";
import CircularProgress from "@mui/material/CircularProgress";
import Link from "@mui/material/Link";
import {
  Alert,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@mui/material";
import { HackInfoResponse, HackOverview } from "../proto/library_checker";
import { RpcError } from "@protobuf-ts/runtime-rpc";
import MainContainer from "../components/MainContainer";
import NotFound from "./NotFound";

const HackInfo: React.FC = () => {
  const { id } = useParams<"id">();
  if (!id) {
    throw new Error(`hack ID is not defined`);
  }
  const intID = parseInt(id);
  if (Number.isNaN(intID)) {
    return <NotFound />;
  }
  return (
    <MainContainer title={`Hack #{id}`}>
      <HackInfoBody id={intID} />
    </MainContainer>
  );
};
export default HackInfo;

const HackInfoBody: React.FC<{
  id: number;
}> = (props) => {
  const { id } = props;
  const hackInfoQuery = useHackInfo(id, {
    refetchInterval: 1000,
  });

  if (hackInfoQuery.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (hackInfoQuery.isError) {
    return (
      <Box>
        <Alert severity="error">
          {(hackInfoQuery.error as RpcError).toString()}
        </Alert>
      </Box>
    );
  }

  const info = hackInfoQuery.data;

  return (
    <Box>
      {info.overview && <OverView overview={info.overview} />}
      <TestCase info={info} />
      {info.stderr && (
        <Box>
          <Typography>Stderr</Typography>
          <pre>{new TextDecoder().decode(info.stderr)}</pre>
        </Box>
      )}
      {info.judgeOutput && (
        <Box>
          <Typography>Judge output</Typography>
          <pre>{new TextDecoder().decode(info.judgeOutput)}</pre>
        </Box>
      )}
    </Box>
  );
};

const OverView: React.FC<{
  overview: HackOverview;
}> = (props) => {
  const { overview } = props;

  return (
    <TableContainer>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Submission</TableCell>
            <TableCell>Hacker</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Time</TableCell>
            <TableCell>Memory</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow>
            <TableCell>
              <Link href={`/submission/${overview.submissionId}`}>
                #{overview.submissionId}
              </Link>
            </TableCell>
            <TableCell>{overview.userName ?? "(Anonymous)"}</TableCell>
            <TableCell>{overview.status}</TableCell>
            <TableCell>
              {overview.time ? Math.round(overview.time * 1000) : "-"} ms
            </TableCell>
            <TableCell>
              {overview.memory
                ? (Number(overview.memory) / 1024 / 1024).toFixed(2)
                : "-"}{" "}
              Mib
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  );
};

const TestCase: React.FC<{
  info: HackInfoResponse;
}> = (props) => {
  const { info } = props;
  const testCase = info.testCase;

  if (testCase.oneofKind === "txt") {
    return (
      <>
        <Typography>TestCase</Typography>
        <pre>{new TextDecoder().decode(testCase.txt)}</pre>
      </>
    );
  } else if (testCase.oneofKind === "cpp") {
    return (
      <>
        <Typography>TestCase</Typography>
        <pre>{new TextDecoder().decode(testCase.cpp)}</pre>
      </>
    );
  } else {
    return (
      <>
        <Typography>TestCase</Typography>
        <pre>not found</pre>
      </>
    );
  }
};
