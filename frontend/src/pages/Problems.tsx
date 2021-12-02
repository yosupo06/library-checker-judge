import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import Typography from "@mui/material/Typography";
import Alert from "@mui/material/Alert";
import Link from "@mui/material/Link";
import React, { useContext, useState } from "react";
import {
  useProblemCategories,
  useProblemList,
  useUserInfo,
} from "../api/library_checker_client";
import { SolvedStatus } from "../api/library_checker_pb";
import ProblemList from "../components/ProblemList";
import { AuthContext } from "../contexts/AuthContext";
import {
  CategorisedProblems,
  categoriseProblems,
} from "../utils/ProblemCategorizer";
import { Tab, Tabs } from "@mui/material";

const ProblemsTabs: React.FC<{
  categories: CategorisedProblems;
  solvedStatus: { [problem: string]: "latest_ac" | "ac" };
}> = (props) => {
  const { categories, solvedStatus } = props;
  const [selectedIdx, setSelectedIdx] = useState(0);

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

const Problems: React.FC = () => {
  const auth = useContext(AuthContext);

  const problemListQuery = useProblemList();
  const problemCategoriesQuery = useProblemCategories();
  const userName = auth?.state.user ?? "";
  const userInfoQuery = useUserInfo(userName, {
    enabled: userName !== "",
  });

  if (
    problemListQuery.isLoading ||
    problemListQuery.isIdle ||
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

  const categories = categoriseProblems(
    problemList,
    problemCategoriesQuery.data.getCategoriesList()
  );

  return (
    <Box>
      <Box>
        <Typography variant="h2" paragraph={true}>
          Problem List
        </Typography>
        <Alert severity="info">
          If you have some trouble, please use{" "}
          <Link href="https://old.yosupo.jp">old.yosupo.jp</Link>
        </Alert>
        <ProblemsTabs categories={categories} solvedStatus={solvedStatus} />
      </Box>
    </Box>
  );
};

export default Problems;
