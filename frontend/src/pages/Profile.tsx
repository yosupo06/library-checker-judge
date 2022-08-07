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
  useUserInfo,
} from "../api/library_checker_client";
import { ChangeUserInfoRequest } from "../api/library_checker_pb";
import { useParams } from "react-router-dom";
import { AuthContext } from "../contexts/AuthContext";
import NotFound from "./NotFound";
import { LibraryBooks } from "@mui/icons-material";
import { Container, FormLabel, Switch } from "@mui/material";
import BuildIcon from "@mui/icons-material/Build";

const Profile: React.FC = () => {
  const { userId } = useParams<"userId">();
  if (!userId) {
    throw new Error(`userId is not passed`);
  }
  const auth = useContext(AuthContext);
  const currentUser = auth?.state.user;

  const [libraryURL, setLibraryURL] = useState("");
  const [isDeveloper, setIsDeveloper] = useState(false);
  const userInfoQuery = useUserInfo(userId, {
    onSuccess: (data) => {
      console.log(data.getUser());
      setLibraryURL(data.getUser()?.getLibraryUrl() ?? "");
      setIsDeveloper(data.getUser()?.getIsDeveloper() ?? false);
    },
  });

  if (userInfoQuery.isLoading || userInfoQuery.isIdle) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (userInfoQuery.isError) {
    return <NotFound />;
  }

  const showUser = userInfoQuery.data.getUser();
  if (!showUser) {
    return <NotFound />;
  }
  const showUserName = showUser.getName();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const newUser = showUser
      .setLibraryUrl(libraryURL)
      .setIsDeveloper(isDeveloper);

    console.log(newUser);

    library_checker_client
      .changeUserInfo(
        new ChangeUserInfoRequest().setUser(newUser),
        (auth && authMetadata(auth.state)) ?? null
      )
      .then(() => {
        history.go(0);
      });
  };

  const form = (
    <Box>
      <Divider />
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
            <FormLabel id="is-developer-mode-switch">Developer Mode</FormLabel>
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
  );

  return (
    <Container>
      <Typography variant="h2">{showUserName}</Typography>
      <List>
        <ListItem>
          <ListItemAvatar>
            <Avatar>
              <LibraryBooks />
            </Avatar>
          </ListItemAvatar>
          <ListItemText
            primary="Library"
            secondary={
              libraryURL ? (
                <Link href={libraryURL}> {libraryURL}</Link>
              ) : (
                "Unregistered"
              )
            }
          />
        </ListItem>
      </List>
      {currentUser && showUserName === currentUser && form}
    </Container>
  );
};

export default Profile;
