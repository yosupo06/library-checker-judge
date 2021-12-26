import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import React from "react";
import RankingList from "../components/RankingList";

interface Props {}

const Ranking: React.FC<Props> = () => {
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
