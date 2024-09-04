import Container, { ContainerProps } from "@mui/material/Container";
import React, { useEffect } from "react";
import KatexTypography from "./katex/KatexTypography";

interface MainContainerProps extends ContainerProps {
  title: string;
}

const MainContainer: React.FC<MainContainerProps> = (props) => {
  // TODO: <title> will be enough in near future: https://react.dev/reference/react-dom/components/title
  useEffect(() => {
    document.title = props.title;
  }, []);
  return (
    <>
      <Container sx={{ paddingBottom: 3 }}>
        <KatexTypography variant="h2">{props.title}</KatexTypography>
        {props.children}
      </Container>
    </>
  );
};

export default MainContainer;
