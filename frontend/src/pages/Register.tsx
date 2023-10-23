import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Container from "@mui/material/Container";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";
import React, { useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import library_checker_client, { useCurrentUser, useRegister } from "../api/client_wrapper";
import { AuthContext } from "../contexts/AuthContext";
import { useMutation } from "@tanstack/react-query";
import { createUserWithEmailAndPassword } from "firebase/auth";
import { useAuth, useCurrentAuthUser } from "../auth/auth";
import { Step, StepContent, StepLabel, Stepper } from "@mui/material";
import { Link } from "react-router-dom";

const RegisterAuth: React.FC<{
}> = () => {
  const auth = useAuth()

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")

  const currentAuthUser = useCurrentAuthUser()

  const mutation = useMutation(() => {
    return createUserWithEmailAndPassword(
      auth,
      email,
      password,
    )
  })
  const onRegister = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate()
  }

  if (currentAuthUser.isLoading || currentAuthUser.isError) {
    return (
      <>
        <Typography>Loading</Typography>
      </>
    )
  }

  if (currentAuthUser.data != null) {
    return (
      <>
        <Typography>Finished</Typography>
      </>
    )
  }

  return (
    <>
      <Typography>Register email</Typography>
      {mutation.isError && (
        <Alert severity="error">{(mutation.error as Error).message}</Alert>
      )}
      <form onSubmit={onRegister}>
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
          Register
        </Button>
      </form>
    </>
  )
}

const RegisterUserID: React.FC<{
}> = () => {
  const [userName, setUserName] = useState("")

  const mutation = useRegister()
  const onRegister = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate(userName)
  }

  return (
    <>
      <Typography>Register user ID</Typography>
      {mutation.isError && (
        <Alert severity="error">{(mutation.error as Error).message}</Alert>
      )}
      <form onSubmit={onRegister}>
        <div>
          <TextField
            required
            label="UserName"
            value={userName}
            onChange={(e) => setUserName(e.target.value)}
          />
        </div>
        <Button color="primary" type="submit">
          Register
        </Button>
      </form>
    </>
  )
}



const Register: React.FC = () => {
  const currentAuthUser = useCurrentAuthUser()
  const currentUser = useCurrentUser()

  let step = 0;
  if (currentAuthUser.data != null) step = 1
  if (currentUser.isSuccess && currentUser.data.user != null) step = 2

  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Register
      </Typography>

      <Stepper activeStep={step} orientation="vertical">
        <Step key={"step1"}>
          <StepLabel>Register email & password</StepLabel>
          <StepContent>
            <RegisterAuth />
          </StepContent>
        </Step>
        <Step key={"step2"}>
          <StepLabel>Register user name</StepLabel>
          <StepContent>
            <RegisterUserID />
          </StepContent>
        </Step>
        <Step key={"step3"}>
          <StepLabel>Finish</StepLabel>
          <StepContent>
            <Link to="/">
              Go to Top Page
            </Link>
          </StepContent>
        </Step>
      </Stepper>
    </Container>
  );
};

export default Register;
