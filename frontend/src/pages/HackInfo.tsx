import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import React from "react";
import { useParams } from "react-router-dom";
import { useHackInfo } from "../api/client_wrapper";
import CircularProgress from "@mui/material/CircularProgress";
import Link from "@mui/material/Link";
import { Alert, Container } from "@mui/material";
import { HackInfoResponse } from "../proto/library_checker";
import { RpcError } from "@protobuf-ts/runtime-rpc";

const HackInfo: React.FC = () => {
  const { ID } = useParams<"ID">();
  if (!ID) {
    throw new Error(`hack ID is not defined`);
  }
  const hackInfoQuery = useHackInfo(parseInt(ID), {
    refetchInterval: 1000,
  });

  if (hackInfoQuery.isLoading) {
    return (
      <Container>
        <CircularProgress />
      </Container>
    );
  }

  if (hackInfoQuery.isError) {
    return (
      <Container>
        <Alert severity="error">
          {(hackInfoQuery.error as RpcError).toString()}
        </Alert>
      </Container>
    );
  }
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Hack #{ID}
      </Typography>
      <HackInfoBody info={hackInfoQuery.data} />
    </Container>
  );
};

export default HackInfo;

const HackInfoBody: React.FC<{
  info: HackInfoResponse;
}> = (props) => {
  const { info } = props;
  return (
    <Box>
      <Typography>
        Submission:{" "}
        <Link href={`/submission/${info.overview?.submissionId}`}>
          #{info.overview?.submissionId}
        </Link>
      </Typography>
      <Typography>Status: {info.overview?.status}</Typography>
      <Typography>TestCase</Typography>
      <pre>{new TextDecoder().decode(info.testCase)}</pre>
      <Typography>Checker output</Typography>
      <pre>{new TextDecoder().decode(info.checkerOut)}</pre>
    </Box>
  );
};
