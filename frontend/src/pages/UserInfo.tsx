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
} from "../api/client_wrapper";
import { useParams } from "react-router-dom";
import { AuthContext } from "../contexts/AuthContext";
import NotFound from "./NotFound";
import { LibraryBooks } from "@mui/icons-material";
import { Container, FormLabel, Switch } from "@mui/material";
import BuildIcon from "@mui/icons-material/Build";
import { useCurrentAuthUser } from "../auth/auth";

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
              showUser.libraryUrl ? (
                <Link href={libraryURL}> {libraryURL}</Link>
              ) : (
                "Unregistered"
              )
            }
          />
        </ListItem>
      </List>
    </Container>
  );
};

export default UserInfo;
