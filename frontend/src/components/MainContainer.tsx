import Container, { ContainerProps } from "@mui/material/Container";
import React from "react";
import KatexTypography from "./katex/KatexTypography";

interface MainContainerProps extends ContainerProps {
  title: string;
}

const MainContainer: React.FC<MainContainerProps> = (props) => (
  <Container sx={{ paddingBottom: 3 }}>
    <KatexTypography variant="h2">{props.title}</KatexTypography>
    {props.children}
  </Container>
);

export default MainContainer;
