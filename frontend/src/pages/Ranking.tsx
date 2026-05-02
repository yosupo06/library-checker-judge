import React from "react";
import RankingList from "../components/RankingList";
import MainContainer from "../components/MainContainer";

const Ranking: React.FC<Record<string, never>> = () => {
  return (
    <MainContainer title="Ranking">
      <RankingList />
    </MainContainer>
  );
};

export default Ranking;
