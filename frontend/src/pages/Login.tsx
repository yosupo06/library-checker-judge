import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";
import Container from "@mui/material/Container";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";
import React from "react";
import { useNavigate } from "react-router-dom";
import {
  useSendPasswordResetEmailMutation,
  useSignInMutation,
} from "../auth/auth";

const Login: React.FC = () => {
  const navigate = useNavigate();

  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");

  const signInMutation = useSignInMutation();
  const onSignIn = (e: React.FormEvent) => {
    e.preventDefault();
    signInMutation.mutate(
      {
        email: email,
        password: password,
      },
      {
        onSuccess: () => navigate(`/`),
      }
    );
  };

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Login
      </Typography>
      <Alert severity="info">
        <AlertTitle>Info</AlertTitle>
        If you regesitered your account without an email, please attach{" "}
        <code>@dummy.judge.yosupo.jp</code> at suffix. <br />
        For example: <code>yosupo</code> →{" "}
        <code>yosupo@dummy.judge.yosupo.jp</code>
      </Alert>
      <form onSubmit={(e) => onSignIn(e)}>
        <div>
          <TextField
            required
            label="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            style={{ width: 300 }}
          />
        </div>
        <div>
          <TextField
            required
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            style={{ width: 300 }}
          />
        </div>
        <Button color="primary" type="submit">
          Login
        </Button>
      </form>
      <PasswordReset />
    </Container>
  );
};

export default Login;

const PasswordReset: React.FC = () => {
  const [email, setEmail] = React.useState("");

  const passwordResetMutation = useSendPasswordResetEmailMutation();
  const onPasswordReset = (e: React.FormEvent) => {
    e.preventDefault();
    passwordResetMutation.mutate(email);
  };

  return (
    <Box>
      <Typography variant="h3" paragraph={true}>
        Password Reset
      </Typography>
      <form onSubmit={(e) => onPasswordReset(e)}>
        <div>
          <TextField
            required
            label="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <Button color="primary" type="submit">
          Send email
        </Button>
      </form>
    </Box>
  );
};
