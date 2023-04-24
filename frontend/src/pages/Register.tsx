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

const Register: React.FC = () => {
  const navigate = useNavigate();
  const auth = useContext(AuthContext);
  const [userName, setUserName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [registerStatus, setRegisterStatus] = React.useState<JSX.Element>(
    <Box />
  );
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setRegisterStatus(<CircularProgress />);
    library_checker_client
      .register({ name: userName, password: password }, {})
      .then((resp) => {
        auth?.dispatch({
          type: "login",
          payload: { token: resp.response.token, user: userName },
        });
        navigate(`/`);
      })
      .catch((reason) =>
        setRegisterStatus(
          <Alert severity="error">
            <AlertTitle>Register failed: {reason.message}</AlertTitle>
          </Alert>
        )
      );
  };
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Register
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
          Register
        </Button>
      </form>
      {registerStatus}
    </Container>
  );
};

export default Register;
