import { Container, Typography } from "@material-ui/core";
import React from "react";
import { connect } from "react-refetch";
import RankingList from "../components/RankingList";

interface Props {}

const Ranking: React.FC<Props> = props => {
  return (
    <Container>
      <Typography variant="h2" paragraph={true}>
        Ranking
      </Typography>
      <RankingList />
    </Container>
  );
};

export default connect<{}, Props>(() => ({}))(Ranking);
