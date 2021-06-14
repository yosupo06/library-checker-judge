import {
  Container,
  createMuiTheme,
  makeStyles,
  ThemeProvider,
  Toolbar,
} from "@material-ui/core";
import React, { useEffect, useReducer } from "react";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import "./App.css";
import NavBar from "./components/NavBar";
import Help from "./pages/Help";
import Login from "./pages/Login";
import ProblemInfo from "./pages/ProblemInfo";
import Problems from "./pages/Problems";
import Ranking from "./pages/Ranking";
import Profile from "./pages/Profile";
import SubmissionInfo from "./pages/SubmissionInfo";
import Submissions from "./pages/Submissions";
import { AuthReducer, AuthContext } from "./contexts/AuthContext";
import { LangReducer, LangContext, LangState } from "./contexts/LangContext";
import { CssBaseline } from "@material-ui/core";
import { grey } from "@material-ui/core/colors";
import NotFound from "./pages/NotFound";
import Register from "./pages/Register";

const theme = createMuiTheme({
  typography: {
    button: {
      textTransform: "none",
    },
  },
  overrides: {
    MuiCssBaseline: {
      "@global": {
        pre: {
          fontFamily: '"Courier New", Consolas, monospace',
          fontSize: "13px",
          background: grey[200],
        },
      },
    },
  },
});

const useStyles = makeStyles((theme) => ({
  root: {
    marginBottom: theme.spacing(4),
  },
}));

function App(): JSX.Element {
  const classes = useStyles();

  const savedLangState = localStorage.getItem("lang");
  let initialLangState: LangState = {
    lang: "en",
  };
  try {
    if (savedLangState) {
      initialLangState = JSON.parse(savedLangState);
    }
  } catch (_) {
    localStorage.removeItem("lang");
  }
  const [langState, langDispatch] = useReducer(LangReducer, initialLangState);
  useEffect(() => {
    localStorage.setItem("lang", JSON.stringify(langState));
  }, [langState]);

  const savedAuthState = localStorage.getItem("auth");
  let initialAuthState = {
    user: "",
    token: "",
  };
  try {
    if (savedAuthState) {
      initialAuthState = JSON.parse(savedAuthState);
    }
  } catch (_) {
    localStorage.removeItem("auth");
  }
  const [authState, authDispatch] = useReducer(AuthReducer, initialAuthState);
  useEffect(() => {
    localStorage.setItem("auth", JSON.stringify(authState));
  }, [authState]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthContext.Provider
        value={{ state: authState, dispatch: authDispatch }}
      >
        <LangContext.Provider
          value={{ state: langState, dispatch: langDispatch }}
        >
          <Router>
            <NavBar />
            <Toolbar />

            <Container className={classes.root}>
              <Switch>
                <Route exact path="/" component={Problems} />
                <Route path="/problem/:problemId" component={ProblemInfo} />
                <Route exact path="/submissions" component={Submissions} />
                <Route
                  path="/submission/:submissionId"
                  component={SubmissionInfo}
                />
                <Route exact path="/ranking" component={Ranking} />
                <Route exact path="/help" component={Help} />
                <Route exact path="/login" component={Login} />
                <Route exact path="/register" component={Register} />
                <Route path="/user/:userId" component={Profile} />
                <Route component={NotFound} />
              </Switch>
            </Container>
          </Router>
        </LangContext.Provider>
      </AuthContext.Provider>
    </ThemeProvider>
  );
}

export default App;
