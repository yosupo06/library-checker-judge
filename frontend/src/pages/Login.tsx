import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Container from "@mui/material/Container";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";
import React, { useContext } from "react";
import { RouteComponentProps } from "react-router-dom";
import library_checker_client from "../api/library_checker_client";
import { LoginRequest } from "../api/library_checker_pb";
import { AuthContext } from "../contexts/AuthContext";

interface Props extends RouteComponentProps<Record<string, never>> {}

const Login: React.FC<Props> = (props) => {
  const { history } = props;
  const auth = useContext(AuthContext);
  const [userName, setUserName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [loginStatus, setLoginStatus] = React.useState<JSX.Element>(<Box />);
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoginStatus(<CircularProgress />);
    library_checker_client
      .login(new LoginRequest().setName(userName).setPassword(password), {})
      .then((resp) => {
        auth?.dispatch({
          type: "login",
          payload: { token: resp.getToken(), user: userName },
        });
        history.push(`/`);
      })
      .catch((reason) =>
        setLoginStatus(
          <Alert severity="error">
            <AlertTitle>Login failed: {reason.message}</AlertTitle>
          </Alert>
        )
      );
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
      {loginStatus}
    </Container>
  );
};

export default Login;
