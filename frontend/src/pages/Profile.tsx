import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Avatar from "@mui/material/Avatar";
import Divider from "@mui/material/Divider";
import Link from "@mui/material/Link";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemAvatar from "@mui/material/ListItemAvatar";
import ListItemText from "@mui/material/ListItemText";
import React, { useContext, useState } from "react";
import library_checker_client, {
  authMetadata,
  useCurrentUser,
  useUserInfo,
} from "../api/client_wrapper";
import { useParams } from "react-router-dom";
import { AuthContext } from "../contexts/AuthContext";
import NotFound from "./NotFound";
import { LibraryBooks } from "@mui/icons-material";
import { Container, FormLabel, Switch } from "@mui/material";
import BuildIcon from "@mui/icons-material/Build";
import { useCurrentAuthUser, useUpdateEmailMutation } from "../auth/auth";
import { User } from "firebase/auth";
import EmailIcon from "@mui/icons-material/Email";

const Profile: React.FC = () => {
  const currentAuthUser = useCurrentAuthUser();
  const currentUser = useCurrentUser();

  const [libraryURL, setLibraryURL] = useState("");
  const [isDeveloper, setIsDeveloper] = useState(false);

  if (currentAuthUser.isLoading || currentUser.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (currentAuthUser.isError || currentUser.isError) {
    // TODO: 500
    return (
      <Box>
        <Typography>Error</Typography>
      </Box>
    );
  }

  const authUser = currentAuthUser.data;
  const user = currentUser.data.user;

  if (!authUser || !user) {
    // TODO: jump to register page?
    return (
      <Box>
        <Typography>Please register</Typography>
      </Box>
    );
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const newUser = user;
    newUser.libraryUrl = libraryURL;
    newUser.isDeveloper = isDeveloper;

    library_checker_client
      .changeUserInfo({ user: newUser }, undefined)
      .then(() => {
        history.go(0);
      });
  };

  return (
    <Container>
      <Typography variant="h2">Profile: {user.name}</Typography>
      <AuthProfile user={authUser} />
      <Box>
        <Typography variant="h4">Setting</Typography>
        <form onSubmit={(e) => handleSubmit(e)}>
          <List>
            <ListItem>
              <ListItemAvatar>
                <Avatar>
                  <LibraryBooks />
                </Avatar>
              </ListItemAvatar>
              <ListItemText
                secondary={
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
              <FormLabel id="is-developer-mode-switch">
                Developer Mode
              </FormLabel>
              <Switch
                aria-labelledby="is-developer-mode-switch"
                checked={isDeveloper}
                onChange={(e) => {
                  console.log(e.target.value, e.target.checked);
                  setIsDeveloper(e.target.checked);
                }}
              />
            </ListItem>
          </List>
          <Button color="primary" type="submit">
            Change
          </Button>
        </form>
      </Box>
    </Container>
  );
};

export default Profile;

const AuthProfile: React.FC<{ user: User }> = (props) => {
  const { user } = props;

  const [newEmail, setNewEmail] = useState(user.email ?? "");

  const updateEmailMutation = useUpdateEmailMutation();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateEmailMutation.mutate(newEmail);
  };

  return (
    <Box>
      <Typography variant="h4">Auth info</Typography>
      <form onSubmit={(e) => handleSubmit(e)}>
        <List>
          <ListItem>
            <ListItemAvatar>
              <Avatar>
                <EmailIcon />
              </Avatar>
            </ListItemAvatar>
            <ListItemText
              secondary={
                <TextField
                  label="Email"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  helperText="Please input URL for your published library"
                />
              }
            />
          </ListItem>
        </List>
        <Button color="primary" type="submit">
          Update email
        </Button>
      </form>
    </Box>
  );
};
