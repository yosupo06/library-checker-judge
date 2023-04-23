import React, { useEffect, useRef } from "react";
import Typography from "@mui/material/Typography";
import renderMathInElement, {
  RenderMathInElementOptions,
} from "katex/contrib/auto-render";

interface Props {
  variant?:
    | "body1"
    | "body2"
    | "button"
    | "caption"
    | "h1"
    | "h2"
    | "h3"
    | "h4"
    | "h5"
    | "h6"
    | "inherit"
    | "overline"
    | "subtitle1"
    | "subtitle2";
  paragraph?: boolean;
}

const renderMathInElementOptions: RenderMathInElementOptions = {
  delimiters: [
    { left: "$$", right: "$$", display: true },
    { left: "\\[", right: "\\]", display: true },
    { left: "$", right: "$", display: false },
    { left: "\\(", right: "\\)", display: false },
  ],
};

const KatexTypography: React.FC<Props> = (props) => {
  const ref = useRef<HTMLElement>(null);
  useEffect(() => {
    if (ref.current) {
      renderMathInElement(ref.current, renderMathInElementOptions);
    }
  });
  return (
    <Typography
      ref={ref}
      children={props.children}
      variant={props.variant}
      paragraph={props.paragraph}
    />
  );
};

export default KatexTypography;
