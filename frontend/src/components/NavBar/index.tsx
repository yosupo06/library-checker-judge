import {
  AppBar,
  Box,
  Button,
  List,
  ListItem,
  ListItemText,
  MenuItem,
  Popover,
  Select,
  Toolbar,
  Typography
} from "@material-ui/core";
import { GitHub } from "@material-ui/icons";
import React, { useContext } from "react";
import { RouteComponentProps, withRouter } from "react-router-dom";
import { AuthContext } from "../../contexts/AuthContext";
import { LangContext } from "../../contexts/LangContext";
import flagJA from "./flag_ja.svg";
import flagUS from "./flag_us.svg";

const NavBar = (props: RouteComponentProps) => {
  const { history } = props;
  const lang = useContext(LangContext);
  const auth = useContext(AuthContext);

  const langSelect = (
    <Select
      value={lang?.state.lang}
      variant="outlined"
      onChange={e =>
        lang?.dispatch({
          type: "change",
          payload: (e.target.value as string) === "ja" ? "ja" : "en"
        })
      }
      style={{
        marginLeft: "auto"
      }}
    >
      <MenuItem value="en">
        <img src={flagUS} alt="us" height="20px"></img>
      </MenuItem>
      <MenuItem value="ja">
        <img src={flagJA} alt="ja" height="20px"></img>
      </MenuItem>
    </Select>
  );

  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    auth?.dispatch({ type: "logout" });
    handleClose();
  };

  const userMenu = (() => {
    if (!auth || !auth.state.user) {
      return (
        <Button color="inherit" onClick={() => history.push("/login")}>
          Login
        </Button>
      );
    }
    return (
      <Box>
        <Button color="inherit" onClick={handleClick}>
          {auth.state.user}
        </Button>
        <Popover
          anchorEl={anchorEl}
          anchorOrigin={{
            vertical: "bottom",
            horizontal: "center"
          }}
          transformOrigin={{
            vertical: "top",
            horizontal: "center"
          }}
          open={Boolean(anchorEl)}
          onClose={handleClose}
        >
          <MenuItem onClick={handleLogout}>Logout</MenuItem>
        </Popover>
      </Box>
    );
  })();

  return (
    <AppBar position="static">
      <Toolbar>
        <List>
          <ListItem>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push("/")}>
                  Library-Checker
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText inset>
              <Button
                color="inherit"
                onClick={() => history.push("/submissions")}
              >
                Submissions
              </Button>
            </ListItemText>
            <ListItemText>
              <Button color="inherit" onClick={() => history.push("/ranking")}>
                Ranking
              </Button>
            </ListItemText>
            <ListItemText>
              <Button color="inherit" onClick={() => history.push("/help")}>
                Help
              </Button>
            </ListItemText>
          </ListItem>
        </List>
        {langSelect}
        {userMenu}
        <Button
          color="inherit"
          href="https://github.com/yosupo06/library-checker-problems"
          target="_blank"
          rel="noopener"
        >
          <GitHub />
        </Button>
      </Toolbar>
    </AppBar>
  );
};

export default withRouter(NavBar);
