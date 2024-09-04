import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Avatar from "@mui/material/Avatar";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemAvatar from "@mui/material/ListItemAvatar";
import ListItemText from "@mui/material/ListItemText";
import React, { useState } from "react";
import {
  useChangeCurrentUserInfoMutation,
  useCurrentUser,
} from "../api/client_wrapper";
import { LibraryBooks } from "@mui/icons-material";
import { Alert, Container, Divider, FormLabel, Switch } from "@mui/material";
import BuildIcon from "@mui/icons-material/Build";
import { useCurrentAuthUser, useUpdateEmailMutation } from "../auth/auth";
import { User as AuthUser } from "firebase/auth";
import EmailIcon from "@mui/icons-material/Email";
import { Navigate } from "react-router-dom";
import { RpcError } from "@protobuf-ts/runtime-rpc";
import { User } from "../proto/library_checker";
import MainContainer from "../components/MainContainer";

const Profile: React.FC = () => {
  const currentAuthUser = useCurrentAuthUser();
  const currentUser = useCurrentUser();

  if (currentAuthUser.isPending || currentUser.isPending) {
    return (
      <Container>
        <CircularProgress />
      </Container>
    );
  }

  if (currentAuthUser.isError || currentUser.isError) {
    return (
      <>
        {currentAuthUser.isError && (
          <Container>
            <Alert severity="error">
              {(currentAuthUser.error as RpcError).toString()}
            </Alert>
          </Container>
        )}
        {currentUser.isError && (
          <Container>
            <Alert severity="error">
              {(currentUser.error as RpcError).toString()}
            </Alert>
          </Container>
        )}
      </>
    );
  }

  const authUser = currentAuthUser.data;
  const user = currentUser.data.user;

  if (!authUser || !user) {
    return <Navigate to={`/register`} />;
  }

  return (
    <MainContainer title={`Profile: ${user.name}`}>
      <GeneralSetting user={user} />
      <Divider
        sx={{
          margin: 1,
        }}
      />
      <EmailSetting user={authUser} />
    </MainContainer>
  );
};
export default Profile;

const EmailSetting: React.FC<{ user: AuthUser }> = (props) => {
  const { user } = props;

  const [newEmail, setNewEmail] = useState(user.email ?? "");

  const mutation = useUpdateEmailMutation();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate(newEmail);
  };

  return (
    <Box>
      <Typography variant="h4">Email</Typography>
      {mutation.isSuccess && (
        <Alert severity="success">Verification email has been sent</Alert>
      )}
      {mutation.isError && (
        <Alert severity="error">{(mutation.error as RpcError).message}</Alert>
      )}
      <form onSubmit={handleSubmit}>
        <List>
          <ListItem>
            <ListItemAvatar>
              <Avatar>
                <EmailIcon />
              </Avatar>
            </ListItemAvatar>
            <ListItemText
              primary={
                <>
                  <TextField
                    label="Email"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                  />
                </>
              }
              secondary={
                <>
                  {user.emailVerified && "verified"}
                  {!user.emailVerified && "unverified"}
                </>
              }
            />
          </ListItem>
        </List>
        <Button color="primary" type="submit">
          Update
        </Button>
      </form>
    </Box>
  );
};

const GeneralSetting: React.FC<{ user: User }> = (props) => {
  const { user } = props;

  const [libraryURL, setLibraryURL] = useState(user.libraryUrl);
  const [isDeveloper, setIsDeveloper] = useState(user.isDeveloper);

  const mutation = useChangeCurrentUserInfoMutation();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate({
      user: {
        name: user.name,
        libraryUrl: libraryURL,
        isDeveloper: isDeveloper,
      },
    });
  };

  return (
    <Box>
      <Typography variant="h4">General</Typography>
      {mutation.isSuccess && <Alert severity="success">Updated</Alert>}
      {mutation.isError && (
        <Alert severity="error">{(mutation.error as RpcError).message}</Alert>
      )}
      <form onSubmit={handleSubmit}>
        <List>
          <ListItem>
            <ListItemAvatar>
              <Avatar>
                <LibraryBooks />
              </Avatar>
            </ListItemAvatar>
            <ListItemText
              primary={
                <TextField
                  label="Library URL"
                  value={libraryURL}
                  onChange={(e) => setLibraryURL(e.target.value)}
                  helperText="Please input URL for your published library"
                />
              }
            />
          </ListItem>
          <ListItem>
            <ListItemAvatar>
              <Avatar>
                <BuildIcon />
              </Avatar>
            </ListItemAvatar>
            <FormLabel id="is-developer-mode-switch">Developer Mode</FormLabel>
            <Switch
              aria-labelledby="is-developer-mode-switch"
              checked={isDeveloper}
              onChange={(e) => {
                setIsDeveloper(e.target.checked);
              }}
            />
          </ListItem>
        </List>
        <Button color="primary" type="submit">
          Update
        </Button>
      </form>
    </Box>
  );
};
