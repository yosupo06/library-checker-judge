import {
  Box,
  Button,
  CircularProgress,
  Link,
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
} from "../api/library_checker_pb";
import { RouteComponentProps } from "react-router-dom";
import { AuthContext } from "../contexts/AuthContext";

interface OuterProps extends RouteComponentProps<{ userId: string }> {}

interface InnerProps extends OuterProps {
  userInfoFetch: PromiseState<UserInfoResponse>;
}

const Profile: React.FC<InnerProps> = (props) => {
  const { userInfoFetch, history } = props;
  const [libraryURL, setLibraryURL] = useState("");
  const auth = useContext(AuthContext);

  if (userInfoFetch.pending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (userInfoFetch.rejected) {
    return (
      <Box>
        <Typography variant="body1">No user</Typography>
      </Box>
    );
  }

  const userInfo = userInfoFetch.value;
  const user = userInfo.getUser();

  const showUser = user?.getName();

  if (!user || !showUser) {
    return (
      <Box>
        <Typography variant="body1">Invalid Response</Typography>
      </Box>
    );
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
    <form onSubmit={(e) => handleSubmit(e)}>
      <Box>
        <TextField
          required
          label="Library URL"
          value={libraryURL}
          onChange={(e) => setLibraryURL(e.target.value)}
        />
      </Box>
      <Button color="primary" type="submit">
        Change
      </Button>
    </form>
  );

  return (
    <Box>
      <Typography variant="h2">{showUser ?? "???"}</Typography>
      <Typography variant="h4">
        Library:{" "}
        {fetchedURL ? (
          <Link href={fetchedURL}>{fetchedURL}</Link>
        ) : (
          "Unregistered"
        )}{" "}
      </Typography>
      {showUser && currentUser && showUser === currentUser && form}
    </Box>
  );
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
