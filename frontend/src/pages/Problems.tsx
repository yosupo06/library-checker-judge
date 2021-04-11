import {
  Box,
  CircularProgress,
  Link,
  makeStyles,
  Typography,
} from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import React, { useContext } from "react";
import { connect, PromiseState } from "react-refetch";
import library_checker_client from "../api/library_checker_client";
import {
  ProblemInfoRequest,
  ProblemListResponse,
  SolvedStatus,
  UserInfoRequest,
  UserInfoResponse,
} from "../api/library_checker_pb";
import ProblemList from "../components/ProblemList";
import { AuthContext } from "../contexts/AuthContext";
import { getCategories } from "../utils/ProblemCategory";

interface OuterProps {}
interface BridgeProps {
  userName: string;
}

interface InnerProps {
  problemListFetch: PromiseState<ProblemListResponse>;
  userInfoFetch: PromiseState<UserInfoResponse>;
}

const useStyles = makeStyles((theme) => ({
  category: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}));

const InnerProblems: React.FC<InnerProps> = (props) => {
  const { problemListFetch, userInfoFetch } = props;
  const classes = useStyles();

  if (problemListFetch.pending || userInfoFetch.pending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (problemListFetch.rejected || userInfoFetch.rejected) {
    return (
      <Box>
        <Typography variant="body1">Error</Typography>
      </Box>
    );
  }

  const problemList = problemListFetch.value.getProblemsList();

  const solvedStatus: { [problem: string]: "latest_ac" | "ac" } = {};
  userInfoFetch.value.toObject().solvedMapMap.forEach((value) => {
    if (value[1] === SolvedStatus.LATEST_AC) {
      solvedStatus[value[0]] = "latest_ac";
    } else if (value[1] === SolvedStatus.AC) {
      solvedStatus[value[0]] = "ac";
    }
  });

  const categories = getCategories(problemList);

  return (
    <Box>
      <Alert severity="info">
        If you have some trouble, please use{" "}
        <Link href="https://old.yosupo.jp">old.yosupo.jp</Link>
      </Alert>
      {categories.map((category) => (
        <Box className={classes.category}>
          <Typography variant="h3">{category.name}</Typography>
          <ProblemList
            problems={category.problems.map((problem) => ({
              name: problem.name,
              title: problem.title,
            }))}
            solvedStatus={solvedStatus}
          />
        </Box>
      ))}
    </Box>
  );
};

const BridgeProblem = connect<BridgeProps, InnerProps>((props) => ({
  problemListFetch: {
    comparison: null,
    value: () =>
      library_checker_client.problemList(new ProblemInfoRequest(), {}),
  },
  userInfoFetch: {
    comparison: null,
    value: () =>
      library_checker_client.userInfo(
        new UserInfoRequest().setName(props.userName ?? ""),
        {}
      ),
  },
}))(InnerProblems);

const Problems: React.FC<OuterProps> = (props: OuterProps) => {
  const auth = useContext(AuthContext);
  return <BridgeProblem userName={auth?.state.user ?? ""} />;
};

export default Problems;
