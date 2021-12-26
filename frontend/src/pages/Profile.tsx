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
import { RouteComponentProps } from "react-router-dom";
import { AuthContext } from "../contexts/AuthContext";
import NotFound from "./NotFound";
import { LibraryBooks } from "@mui/icons-material";

const Profile: React.FC<RouteComponentProps<{ userId: string }>> = (props) => {
  const { history, match } = props;
  const auth = useContext(AuthContext);
  const [libraryURL, setLibraryURL] = useState("");

  const userName = match.params.userId;
  const userInfoQuery = useUserInfo(userName, {
    onSuccess: (data) => setLibraryURL(data.getUser()?.getLibraryUrl() ?? ""),
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

  const userInfo = userInfoQuery.data;
  const user = userInfo.getUser();

  if (!user) {
    return <NotFound />;
  }

  const showUser = user.getName();

  if (!showUser) {
    return <NotFound />;
  }

  const fetchedURL = user.getLibraryUrl();
  const currentUser = auth?.state.user;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    library_checker_client
      .changeUserInfo(
        new ChangeUserInfoRequest().setUser(
          user.setName(showUser).setLibraryUrl(libraryURL)
        ),
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
        </List>
        <Button color="primary" type="submit">
          Change
        </Button>
      </form>
    </Box>
  );

  return (
    <Box>
      <Typography variant="h2">{showUser ?? "???"}</Typography>
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
              fetchedURL ? (
                <Link href={fetchedURL}> {fetchedURL}</Link>
              ) : (
                "Unregistered"
              )
            }
          />
        </ListItem>
      </List>
      {showUser && currentUser && showUser === currentUser && form}
    </Box>
  );
};

export default Profile;
