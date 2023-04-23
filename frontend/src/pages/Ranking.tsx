import { Container } from "@mui/material";
import Typography from "@mui/material/Typography";
import React from "react";
import RankingList from "../components/RankingList";

const Ranking: React.FC<Record<string, never>> = () => {
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Ranking
      </Typography>
      <RankingList />
    </Container>
  );
};

export default Ranking;
