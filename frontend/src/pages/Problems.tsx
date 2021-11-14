import {
  Box,
  CircularProgress,
  Link,
  makeStyles,
  Typography,
} from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import React, { useContext } from "react";
import { useQuery } from "react-query";
import library_checker_client from "../api/library_checker_client";
import {
  ProblemCategoriesRequest,
  ProblemListRequest,
  SolvedStatus,
  UserInfoRequest,
} from "../api/library_checker_pb";
import ProblemList from "../components/ProblemList";
import { AuthContext } from "../contexts/AuthContext";
import { getCategories } from "../utils/ProblemCategory";

const useStyles = makeStyles((theme) => ({
  category: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}));

const Problems: React.FC = () => {
  const classes = useStyles();
  const auth = useContext(AuthContext);
  const userName = auth?.state.user ?? "";
  const problemListQuery = useQuery("problemList", () =>
    library_checker_client.problemList(new ProblemListRequest(), {})
  );
  const problemCategoriesQuery = useQuery("problemCategories", () =>
    library_checker_client.problemCategories(new ProblemCategoriesRequest(), {})
  );

  const userInfoQuery = useQuery(["userInfo", userName], () =>
    userName
      ? library_checker_client.userInfo(
          new UserInfoRequest().setName(userName),
          {}
        )
      : null
  );

  if (
    problemListQuery.isLoading ||
    problemListQuery.isIdle ||
    userInfoQuery.isLoading ||
    userInfoQuery.isIdle ||
    problemCategoriesQuery.isLoading ||
    problemCategoriesQuery.isIdle
  ) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (
    problemListQuery.isError ||
    userInfoQuery.isError ||
    problemCategoriesQuery.isError
  ) {
    return (
      <Box>
        <Typography variant="body1">
          Error: {problemListQuery.error} {userInfoQuery.error}{" "}
          {problemCategoriesQuery.error}
        </Typography>
      </Box>
    );
  }

  const problemList = problemListQuery.data.getProblemsList();

  const solvedStatus: { [problem: string]: "latest_ac" | "ac" } = {};
  if (userInfoQuery.data != null) {
    userInfoQuery.data.toObject().solvedMapMap.forEach((value) => {
      if (value[1] === SolvedStatus.LATEST_AC) {
        solvedStatus[value[0]] = "latest_ac";
      } else if (value[1] === SolvedStatus.AC) {
        solvedStatus[value[0]] = "ac";
      }
    });
  }

  const categories = getCategories(
    problemList,
    problemCategoriesQuery.data.getCategoriesList()
  );

  return (
    <Box>
      <Alert severity="info">
        If you have some trouble, please use{" "}
        <Link href="https://old.yosupo.jp">old.yosupo.jp</Link>
      </Alert>
      <Box>
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
    </Box>
  );
};

export default Problems;
