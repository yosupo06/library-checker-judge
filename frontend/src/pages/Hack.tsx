import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import SourceEditor from "../components/SourceEditor";
import { useHackMutation } from "../api/client_wrapper";
import { Container, FormControl, TextField } from "@mui/material";
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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate({
      submission: parseInt(submissionId),
      testCase: new TextEncoder().encode(refactorTestCase(testCase)),
    });
  };

  return (
    <Container>
      <Typography variant="h4" paragraph={true}>
        Hack
      </Typography>

      <form onSubmit={handleSubmit}>
        <FormControl>
          <TextField
            label="Submission ID"
            value={submissionId}
            onChange={(e) => setSubmissionId(e.target.value)}
          />
        </FormControl>
        <FormControl
          sx={{
            height: "400px",
            width: "100%",
          }}
        >
          <SourceEditor
            value={testCase}
            language="txt"
            onChange={(e) => {
              setTestCase(e);
            }}
            readOnly={false}
            autoHeight={false}
          />
        </FormControl>
        <Button color="primary" type="submit">
          Hack
        </Button>
      </form>
    </Container>
  );
};

export default Hack;
