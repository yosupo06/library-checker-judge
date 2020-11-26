import {
  Box,
  Card,
  CardContent,
  CircularProgress,
  List,
  ListItem,
  Typography
} from "@material-ui/core";
import React from "react";
import { connect, PromiseState } from "react-refetch";
import { Link } from "react-router-dom";
import library_checker_client from "../api/library_checker_client";
import {
  ProblemInfoRequest,
  ProblemListResponse
} from "../api/library_checker_pb";
import KatexRender from "../components/KatexRender";

interface Props {
  problemListFetch: PromiseState<ProblemListResponse>;
}

const ProblemList: React.FC<Props> = props => {
  const { problemListFetch } = props;

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
  return (
    <Box>
      <List>
        {problemListFetch.value.getProblemsList().map(problem => {
          return (
            <ListItem>
              <Card>
                <CardContent>
                  <Link to={`/problem/${problem.getName()}`}>
                    <KatexRender text={problem.getTitle()} />
                  </Link>
                </CardContent>
              </Card>
            </ListItem>
          );
        })}
      </List>
    </Box>
  );
};

export default connect<{}, Props>(() => ({
  problemListFetch: {
    comparison: null,
    value: library_checker_client.problemList(new ProblemInfoRequest())
  }
}))(ProblemList);
