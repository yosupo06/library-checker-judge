import React from "react";
import { Alert, Box, CircularProgress } from "@mui/material";
import { Lang } from "../contexts/LangContext";
import { StatementData, useStatementParser } from "../utils/statement.parser";
import KatexRender from "../components/katex/KatexRender";
// This component only renders given statement text; fetching is handled elsewhere.

const Statement: React.FC<{
  lang: Lang;
  data: StatementData;
}> = (props) => {
  const statement = useStatementParser(props.lang, props.data);

  if (statement.isPending) {
    return (
      <Box>
        <CircularProgress />
      </Box>
    );
  }

  if (statement.isError) {
    return (
      <>
        <Box>
          <Alert severity="error">{(statement.error as Error).message}</Alert>
        </Box>
      </>
    );
  }

  return (
    <Box>
      <KatexRender text={statement.data} />
    </Box>
  );
};

export default Statement;
