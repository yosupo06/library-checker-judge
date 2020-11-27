import { Box, Typography } from "@material-ui/core";
import React from "react";
import RankingList from "../components/RankingList";

interface Props {}

const Ranking: React.FC<Props> = props => {
  return (
    <Box>
      <Typography variant="h2" paragraph={true}>
        Ranking
      </Typography>
      <RankingList />
    </Box>
  );
};

export default Ranking;
