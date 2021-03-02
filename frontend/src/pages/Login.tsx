import {
  Button,
  CircularProgress,
  Container,
  TextField,
  Typography,
} from "@material-ui/core";
import { Alert, AlertTitle } from "@material-ui/lab";
import React, { useContext } from "react";
import library_checker_client from "../api/library_checker_client";
import { LoginRequest } from "../api/library_checker_pb";
import { AuthContext } from "../contexts/AuthContext";

interface Props {}

const Help: React.FC<Props> = (props) => {
  const auth = useContext(AuthContext);
  const [userName, setUserName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [loginStatus, setLoginStatus] = React.useState<
    "none" | "wait" | "success" | "failed"
  >("none");
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoginStatus("wait");
    library_checker_client
      .login(new LoginRequest().setName(userName).setPassword(password))
      .then((resp) => {
        auth?.dispatch({
          type: "login",
          payload: { token: resp.getToken(), user: userName },
        });
        setLoginStatus("success");
      })
      .catch((reason) => setLoginStatus("failed"));
  };
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Login
      </Typography>
      <form onSubmit={(e) => handleSubmit(e)}>
        <div>
          <TextField
            required
            label="User Name"
            value={userName}
            onChange={(e) => setUserName(e.target.value)}
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
      {loginStatus === "wait" && <CircularProgress />}
      {loginStatus === "success" && (
        <Alert severity="success">
          <AlertTitle>Success</AlertTitle>
          Success: Login
        </Alert>
      )}
      {loginStatus === "failed" && (
        <Alert severity="error">
          <AlertTitle>Failed</AlertTitle>
          Failed: Login
        </Alert>
      )}
    </Container>
  );
};

export default Help;
