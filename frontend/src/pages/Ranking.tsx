import { Container } from "@mui/material";
import Typography from "@mui/material/Typography";
import React from "react";
import RankingList from "../components/RankingList";

interface Props {}

const Ranking: React.FC<Props> = () => {
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
