import Container from "@mui/material/Container";
import Toolbar from "@mui/material/Toolbar";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import { useEffect, useReducer } from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
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
import CssBaseline from "@mui/material/CssBaseline";
import { grey } from "@mui/material/colors";
import NotFound from "./pages/NotFound";
import Register from "./pages/Register";
import { QueryClient, QueryClientProvider } from "react-query";
const theme = createTheme({
  typography: {
    button: {
      textTransform: "none",
    },
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        pre: {
          fontFamily: '"Courier New", Consolas, monospace',
          fontSize: "13px",
          background: grey[200],
        },
      },
    },
  },
});

const queryClient = new QueryClient();

function App(): JSX.Element {
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
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider
          value={{ state: authState, dispatch: authDispatch }}
        >
          <LangContext.Provider
            value={{ state: langState, dispatch: langDispatch }}
          >
            <Router>
              <NavBar />
              <Toolbar />
              <Container>
                <Routes>
                  <Route path="/" element={<Problems />} />
                  <Route path="/problem/:problemId" element={<ProblemInfo />} />
                  <Route path="/submissions" element={<Submissions />} />
                  <Route
                    path="/submission/:submissionId"
                    element={<SubmissionInfo />}
                  />
                  <Route path="/ranking" element={<Ranking />} />
                  <Route path="/help" element={<Help />} />
                  <Route path="/login" element={<Login />} />
                  <Route path="/register" element={<Register />} />
                  <Route path="/user/:userId" element={<Profile />} />
                  <Route element={NotFound} />
                </Routes>
              </Container>
            </Router>
          </LangContext.Provider>
        </AuthContext.Provider>
      </QueryClientProvider>
    </ThemeProvider>
  );
}

export default App;
