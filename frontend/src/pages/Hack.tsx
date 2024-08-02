import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import SourceEditor from "../components/SourceEditor";
import { useHackMutation } from "../api/client_wrapper";
import {
  Box,
  Container,
  FormControl,
  Tab,
  Tabs,
  TextField,
} from "@mui/material";
import { refactorTestCase } from "../utils/hack";

const Hack: React.FC = () => {
  const navigate = useNavigate();
  const mutation = useHackMutation({
    onSuccess: (resp) => {
      navigate(`/hack/${resp.id}`);
    },
  });
  const [submissionId, setSubmissionId] = useState("");
  const [testCase, setTestCase] = useState("");
  const [tabIndex, setTabIndex] = useState(0);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate({
      submission: parseInt(submissionId),
      testCase: new TextEncoder().encode(refactorTestCase(testCase)),
    });
  };

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Hack (β)
      </Typography>

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
              value={testCase}
              onChange={(e) => {
                setTestCase(e);
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
            <Typography variant="h4" paragraph={true}>
              TODO
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
