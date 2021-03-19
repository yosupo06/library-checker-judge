import {
  Avatar,
  Box,
  Button,
  CircularProgress,
  Divider,
  Link,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  makeStyles,
  TextField,
  Typography,
} from "@material-ui/core";
import React, { useContext, useState } from "react";
import { connect, PromiseState } from "react-refetch";
import library_checker_client, {
  authMetadata,
} from "../api/library_checker_client";
import {
  ChangeUserInfoRequest,
  UserInfoRequest,
  UserInfoResponse,
  User,
} from "../api/library_checker_pb";
import { RouteComponentProps } from "react-router-dom";
import { AuthContext } from "../contexts/AuthContext";
import NotFound from "./NotFound";
import LibraryBooksIcon from "@material-ui/icons/LibraryBooks";
import * as H from "history";

interface OuterProps extends RouteComponentProps<{ userId: string }> {}

interface InnerProps extends OuterProps {
  userInfoFetch: PromiseState<UserInfoResponse>;
}

interface InnerProfileProps {
  user: User;
  history: H.History;
}

const useStyles = makeStyles((theme) => ({
  divider: {
    margin: theme.spacing(1),
  },
}));

const InnerProfile: React.FC<InnerProfileProps> = (props) => {
  const { user, history } = props;
  const classes = useStyles();
  const [libraryURL, setLibraryURL] = useState(user.getLibraryUrl());
  const auth = useContext(AuthContext);

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
      .then((resp) => {
        history.go(0);
      });
  };

  const form = (
    <Box>
      <Divider className={classes.divider} />
      <Typography variant="h4">Setting</Typography>
      <form onSubmit={(e) => handleSubmit(e)}>
        <List>
          <ListItem>
            <ListItemAvatar>
              <Avatar>
                <LibraryBooksIcon />
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
              <LibraryBooksIcon />
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

const Profile: React.FC<InnerProps> = (props) => {
  const { userInfoFetch, history } = props;

  if (userInfoFetch.pending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (userInfoFetch.rejected) {
    return <NotFound />;
  }

  const userInfo = userInfoFetch.value;
  const user = userInfo.getUser();

  if (!user) {
    return <NotFound />;
  }

  return <InnerProfile user={user} history={history} />;
};

export default connect<OuterProps, InnerProps>((props) => ({
  userInfoFetch: {
    comparison: null,
    value: () =>
      library_checker_client.userInfo(
        new UserInfoRequest().setName(props.match.params.userId),
        {}
      ),
  },
}))(Profile);
