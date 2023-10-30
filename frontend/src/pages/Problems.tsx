import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import React, { useEffect, useState } from "react";
import {
  useCurrentUser,
  useProblemCategories,
  useProblemList,
  useUserInfo,
} from "../api/client_wrapper";
import { SolvedStatus } from "../proto/library_checker";
import ProblemList from "../components/ProblemList";
import { RpcError } from "@protobuf-ts/runtime-rpc";

import {
  CategorisedProblems,
  categoriseProblems,
} from "../utils/ProblemCategorizer";
import { Alert, Container, Tab, Tabs } from "@mui/material";
import { useLocation, useNavigate } from "react-router-dom";

const Problems: React.FC = () => (
  <Container>
    <Box>
      <Typography variant="h2" paragraph={true}>
        Problem List
      </Typography>
      <ProblemsBody />
    </Box>
  </Container>
);
export default Problems;

const ProblemsBody: React.FC = () => {
  const currentUser = useCurrentUser();
  const userName = currentUser.data?.user?.name ?? "";

  const problemListQuery = useProblemList();
  const problemCategoriesQuery = useProblemCategories();
  const userInfoQuery = useUserInfo(userName, {
    enabled: userName !== "",
  });

  if (problemListQuery.isLoading || problemCategoriesQuery.isLoading) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (problemListQuery.isError || problemCategoriesQuery.isError) {
    return (
      <Box>
        {problemListQuery.isError && (
          <Alert severity="error">
            {(problemListQuery.error as RpcError).toString()}
          </Alert>
        )}
        {problemCategoriesQuery.isError && (
          <Alert severity="error">
            {(problemCategoriesQuery.error as RpcError).toString()}
          </Alert>
        )}
      </Box>
    );
  }

  const problemList = problemListQuery.data.problems;

  const solvedStatus: { [problem: string]: "latest_ac" | "ac" } = {};
  if (userInfoQuery.data != null) {
    Object.entries(userInfoQuery.data.solvedMap).forEach(([p, status]) => {
      if (status === SolvedStatus.LATEST_AC) {
        solvedStatus[p] = "latest_ac";
      } else if (status === SolvedStatus.AC) {
        solvedStatus[p] = "ac";
      }
    });
  }

  const categories = categoriseProblems(
    problemList,
    problemCategoriesQuery.data.categories
  );

  return (
    <Box>
      {userInfoQuery.isError && (
        <Alert severity="error">
          {(userInfoQuery.error as RpcError).toString()}
        </Alert>
      )}
      <ProblemsTabs
        categories={categories}
        solvedStatus={
          userInfoQuery.isSuccess
            ? toSolvedStatus(userInfoQuery.data.solvedMap)
            : undefined
        }
      />
    </Box>
  );
};

const toSolvedStatus = (solvedMap: { [key: string]: SolvedStatus }) => {
  const solvedStatus: { [problem: string]: "latest_ac" | "ac" } = {};
  Object.entries(solvedMap).forEach(([p, status]) => {
    if (status === SolvedStatus.LATEST_AC) {
      solvedStatus[p] = "latest_ac";
    } else if (status === SolvedStatus.AC) {
      solvedStatus[p] = "ac";
    }
  });
  return solvedStatus;
};

type ProblemsTabState = {
  selectedIdx?: number;
} | null;

const ProblemsTabs: React.FC<{
  categories: CategorisedProblems;
  solvedStatus?: { [problem: string]: "latest_ac" | "ac" };
}> = (props) => {
  const { categories, solvedStatus } = props;
  const navigate = useNavigate();
  const location = useLocation();
  const [selectedIdx, setSelectedIdx] = useState(() => {
    const state = location.state as ProblemsTabState;
    return state?.selectedIdx ?? 0;
  });

  useEffect(() => {
    const state = location.state as ProblemsTabState;
    if (state?.selectedIdx !== selectedIdx) {
      navigate(".", { state: { selectedIdx }, replace: true });
    }
  }, [selectedIdx]);
  useEffect(() => {
    const state = location.state as ProblemsTabState;
    if (state?.selectedIdx) {
      setSelectedIdx(state.selectedIdx);
    }
  }, [location.state]);

  const categoriesTab = (
    <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
      <Tabs
        value={selectedIdx}
        onChange={(_, newValue: number) => {
          setSelectedIdx(newValue);
        }}
        variant="scrollable"
        scrollButtons="auto"
      >
        <Tab id="All" label="All" />
        {categories.map((category) => (
          <Tab id={category.name} label={category.name} />
        ))}
      </Tabs>
    </Box>
  );

  const targetCategories =
    selectedIdx === 0 ? categories : [categories[selectedIdx - 1]];
  return (
    <Box>
      {categoriesTab}
      {targetCategories.map((category) => (
        <Box
          sx={{
            marginTop: 3,
            marginBottom: 3,
          }}
        >
          <Typography variant="h4">{category.name}</Typography>
          <ProblemList
            problems={category.problems}
            solvedStatus={solvedStatus}
          />
        </Box>
      ))}
    </Box>
  );
};
