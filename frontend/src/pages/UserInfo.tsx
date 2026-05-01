import CircularProgress from "@mui/material/CircularProgress";
import Avatar from "@mui/material/Avatar";
import Link from "@mui/material/Link";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemAvatar from "@mui/material/ListItemAvatar";
import ListItemText from "@mui/material/ListItemText";
import React from "react";
import { useUserInfo } from "../api/client_wrapper";
import { useParams } from "react-router-dom";
import { LibraryBooks } from "@mui/icons-material";
import MainContainer from "../components/MainContainer";
import { Alert } from "@mui/material";

const UserInfo: React.FC = () => {
  const { userId } = useParams<"userId">();
  if (!userId) {
    throw new Error(`userId is not passed`);
  }

  return (
    <MainContainer title={userId}>
      <UserInfoBody id={userId} />
    </MainContainer>
  );
};

export default UserInfo;

const UserInfoBody: React.FC<{ id: string }> = (props) => {
  const { id } = props;

  const userInfoQuery = useUserInfo(id);

  if (userInfoQuery.isPending) {
    return (
      <>
        <CircularProgress />
      </>
    );
  }
  if (userInfoQuery.isError) {
    return (
      <>
        <Alert severity="error">{userInfoQuery.error.message}</Alert>
      </>
    );
  }

  const user = userInfoQuery.data.user;
  if (!user) {
    return (
      <>
        <Alert severity="warning">User {id} is not found</Alert>
      </>
    );
  }

  return (
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
            user.libraryUrl ? (
              <Link href={user.libraryUrl}> {user.libraryUrl}</Link>
            ) : (
              "Unregistered"
            )
          }
        />
      </ListItem>
    </List>
  );
};
