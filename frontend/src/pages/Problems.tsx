import { Box, CircularProgress, Link, makeStyles, Typography } from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import library_checker_client from "../api/library_checker_client";
import {
  ProblemInfoRequest,
  ProblemListResponse
} from "../api/library_checker_pb";
import ProblemList from "../components/ProblemList";
import { getCategories } from "../utils/ProblemCategory";

interface Props {
  problemListFetch: PromiseState<ProblemListResponse>;
}

const useStyles = makeStyles(theme => ({
  category: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2)
  }
}));

const Problems: React.FC<Props> = props => {
  const { problemListFetch } = props;
  const classes = useStyles()

  if (problemListFetch.pending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }
  if (problemListFetch.rejected) {
    return (
      <Box>
        <Typography variant="body1">Error</Typography>
      </Box>
    );
  }

  const problemList = problemListFetch.value.getProblemsList()

  const categories = getCategories(problemList)

  return (
    <Box>
      <Alert severity="info">If you have some trouble, please use <Link href="https://old.yosupo.jp">old.yosupo.jp</Link></Alert>
      {categories.map(category => (
        <Box className={classes.category}>
          <Typography variant="h3">{category.name}</Typography>
          <ProblemList
            problems={category.problems.map(problem => ({
              name: problem.name,
              title: problem.title,
            }))}
          />
        </Box>
      ))}
    </Box>
  );
};

export default connect<{}, Props>(() => ({
  problemListFetch: {
    comparison: null,
    value: () => library_checker_client.problemList(new ProblemInfoRequest())
  }
}))(Problems);
