import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import Avatar from "@mui/material/Avatar";
import Link from "@mui/material/Link";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemAvatar from "@mui/material/ListItemAvatar";
import ListItemText from "@mui/material/ListItemText";
import React, { useState } from "react";
import { useUserInfo } from "../api/client_wrapper";
import { useParams } from "react-router-dom";
import NotFound from "./NotFound";
import { LibraryBooks } from "@mui/icons-material";
import MainContainer from "../components/MainContainer";

const UserInfo: React.FC = () => {
  const { userId } = useParams<"userId">();
  if (!userId) {
    throw new Error(`userId is not passed`);
  }

  const [libraryURL, setLibraryURL] = useState("");
  const userInfoQuery = useUserInfo(userId, {
    onSuccess: (data) => {
      setLibraryURL(data.user?.libraryUrl ?? "");
    },
  });

  if (userInfoQuery.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (userInfoQuery.isError) {
    return <NotFound />;
  }

  const showUser = userInfoQuery.data.user;
  if (!showUser) {
    return <NotFound />;
  }
  const showUserName = showUser.name;

  return (
    <MainContainer title={showUserName}>
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
              showUser.libraryUrl ? (
                <Link href={libraryURL}> {libraryURL}</Link>
              ) : (
                "Unregistered"
              )
            }
          />
        </ListItem>
      </List>
    </MainContainer>
  );
};

export default UserInfo;
