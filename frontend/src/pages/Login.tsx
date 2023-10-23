import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Container from "@mui/material/Container";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";
import React, { useContext } from "react";
import { useNavigate } from "react-router-dom";
import library_checker_client from "../api/client_wrapper";
import { AuthContext } from "../contexts/AuthContext";
import { useSignInMutation } from "../auth/auth";

const Login: React.FC = () => {
  const navigate = useNavigate();

  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");

  const signInMutation = useSignInMutation()

  const onSignIn = (e: React.FormEvent) => {
    e.preventDefault();
    console.log(email, password)
    signInMutation.mutate({
      email: email,
      password: password,
    })
  };

  console.log(signInMutation)

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Login
      </Typography>
      <form onSubmit={(e) => onSignIn(e)}>
        <div>
          <TextField
            required
            label="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div>
          <TextField
            required
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <Button color="primary" type="submit">
          Login
        </Button>
      </form>
    </Container>
  );
};

export default Login;
