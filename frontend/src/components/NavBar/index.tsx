import {
  AppBar,
  Box,
  Button,
  List,
  ListItem,
  ListItemText,
  makeStyles,
  MenuItem,
  Popover,
  Select,
  Toolbar,
  Typography
} from "@material-ui/core";
import { GitHub } from "@material-ui/icons";
import React, { useContext } from "react";
import { Link, RouteComponentProps, withRouter } from "react-router-dom";
import { AuthContext } from "../../contexts/AuthContext";
import { LangContext } from "../../contexts/LangContext";
import flagJA from "./flag_ja.svg";
import flagUS from "./flag_us.svg";

const useStyles = makeStyles(theme => ({
  langSelect: {
    marginLeft: "auto",
    marginRight: theme.spacing(0.2)
  },
  navbarItem: {
    minWidth: 0,
    marginLeft: theme.spacing(0.2),
    marginRight: theme.spacing(0.2)
  },
  navbarTop: {
    minWidth: 0,
    marginLeft: theme.spacing(0.4),
    marginRight: theme.spacing(0.4)
  },
  navbarTopLink: {
    color: " inherit",
    textDecoration: "none",
    fontSize: "16px",
    fontWeight: "bolder"
  },
  navbarLink: {
    color: " inherit",
    textDecoration: "none"
  }
}));

const NavBar = (props: RouteComponentProps) => {
  const { history } = props;
  const lang = useContext(LangContext);
  const auth = useContext(AuthContext);
  const classes = useStyles();

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
      className={classes.langSelect}
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
              <Button color="inherit" className={classes.navbarTop}>
                <Link to="/" className={classes.navbarTopLink}>
                  Library-Checker
                </Link>
              </Button>
            </ListItemText>
            <ListItemText>
              <Button color="inherit" className={classes.navbarItem}>
                <Link to="/submissions" className={classes.navbarLink}>
                  Submissions
                </Link>
              </Button>
            </ListItemText>
            <ListItemText>
              <Button color="inherit" className={classes.navbarItem}>
                <Link to="/ranking" className={classes.navbarLink}>
                  Ranking
                </Link>
              </Button>
            </ListItemText>
            <ListItemText>
              <Button color="inherit" className={classes.navbarItem}>
                <Link to="/help" className={classes.navbarLink}>
                  Help
                </Link>
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
