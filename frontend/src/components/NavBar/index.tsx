import {
  AppBar,
  Button,
  List,
  ListItem,
  ListItemText,
  MenuItem,
  Select,
  Toolbar,
  Typography
} from "@material-ui/core";
import React, { useContext } from "react";
import { RouteComponentProps, withRouter } from "react-router-dom";
import { LangContext } from "../../contexts/LangContext";
import { AuthContext } from "../../contexts/AuthContext";
import { GitHub } from "@material-ui/icons";
import flagUS from "./flag_us.svg";
import flagJA from "./flag_ja.svg";

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
  return (
    <AppBar position="static">
      <Toolbar>
        <List>
          <ListItem>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push("/")}>
                  LIBRARY-CHECKER
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText inset>
              <Typography color="inherit" variant="h6">
                <Button
                  color="inherit"
                  onClick={() => history.push("/submissions")}
                >
                  SUBMISSIONS
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button
                  color="inherit"
                  onClick={() => history.push("/ranking")}
                >
                  RANKING
                </Button>
              </Typography>
            </ListItemText>
            <ListItemText>
              <Typography color="inherit" variant="h6">
                <Button color="inherit" onClick={() => history.push("/help")}>
                  HELP
                </Button>
              </Typography>
            </ListItemText>
          </ListItem>
        </List>
        {langSelect}
        {!auth?.state.user && (
          <Button color="inherit" onClick={() => history.push("/login")}>
            Login
          </Button>
        )}
        {auth?.state.user && (
          <Typography color="inherit" variant="h6">
            {auth?.state.user}
          </Typography>
        )}
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
