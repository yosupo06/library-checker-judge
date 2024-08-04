import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import SourceEditor from "../components/SourceEditor";
import { useHackMutation } from "../api/client_wrapper";
import {
  Alert,
  Box,
  Container,
  FormControl,
  Tab,
  Tabs,
  TextField,
} from "@mui/material";
import { refactorTestCase } from "../utils/hack";
import { RpcError } from "@protobuf-ts/runtime-rpc";

const Hack: React.FC = () => {
  const navigate = useNavigate();
  const mutation = useHackMutation({
    onSuccess: (resp) => {
      navigate(`/hack/${resp.id}`);
    },
  });
  const [submissionId, setSubmissionId] = useState("");
  const [testCaseTxt, setTestCaseTxt] = useState("");
  const [testCaseCpp, setTestCaseCpp] = useState("");
  const [tabIndex, setTabIndex] = useState(0);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (tabIndex === 0) {
      mutation.mutate({
        submission: parseInt(submissionId),
        testCase: {
          oneofKind: "txt",
          txt: new TextEncoder().encode(refactorTestCase(testCaseTxt)),
        },
      });
    } else {
      mutation.mutate({
        submission: parseInt(submissionId),
        testCase: {
          oneofKind: "cpp",
          cpp: new TextEncoder().encode(refactorTestCase(testCaseCpp)),
        },
      });
    }
  };
  
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Hack (β)
      </Typography>

      {mutation.isSuccess && <Alert severity="success">Hack submitted</Alert>}
      {mutation.isError && (
        <Alert severity="error">{(mutation.error as RpcError).message}</Alert>
      )}

      <Box component="form" onSubmit={handleSubmit}>
        <FormControl>
          <TextField
            label="Submission ID"
            value={submissionId}
            onChange={(e) => setSubmissionId(e.target.value)}
          />
        </FormControl>
        <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Tabs
            value={tabIndex}
            onChange={(_, newValue) => setTabIndex(newValue)}
          >
            <Tab label="Text" />
            <Tab label="Generator" />
          </Tabs>
        </Box>

        {tabIndex === 0 && (
          <Box sx={{ p: 3 }}>
            <SourceEditor
              value={testCaseTxt}
              onChange={(e) => {
                setTestCaseTxt(e);
              }}
              readOnly={false}
              height={600}
            />
            <Typography variant="caption">
              Max length: 1MiB(=2<sup>20</sup>Byte)
            </Typography>
          </Box>
        )}
        {tabIndex === 1 && (
          <Box sx={{ p: 3 }}>
            <SourceEditor
              value={testCaseCpp}
              onChange={(e) => {
                setTestCaseCpp(e);
              }}
              readOnly={false}
              language="cpp"
              height={600}
            />
            <Typography variant="caption">
              Max length: 1MiB(=2<sup>20</sup>Byte)
            </Typography>
          </Box>
        )}
        <Button color="primary" type="submit">
          Hack
        </Button>
      </Box>
    </Container>
  );
};

export default Hack;
