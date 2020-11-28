import renderMathInElement from "katex/dist/contrib/auto-render";
import "katex/dist/katex.min.css";
import React, { useContext } from "react";
import { LangContext } from "../contexts/LangContext";

interface Props {
  text: string;
  html?: boolean;
}

const KatexRender: React.FC<Props> = props => {
  const lang = useContext(LangContext);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (ref.current != null) {
      renderMathInElement(ref.current, {
        delimiters: [
          { left: "$$", right: "$$", display: true },
          { left: "\\[", right: "\\]", display: true },
          { left: "$", right: "$", display: false },
          { left: "\\(", right: "\\)", display: false }
        ],
        ignoredTags: ["script", "noscript", "style"]
      });
      const elems = Array.from(
        ref.current?.getElementsByClassName("lang-ja")
      ).concat(Array.from(ref.current?.getElementsByClassName("lang-en")));
      elems.forEach(e => {
        if (e.classList.contains(`lang-${lang?.state.lang}`)) {
          e.removeAttribute("hidden");
        } else {
          e.setAttribute("hidden", "true");
        }
      });
    }
  });

  if (props.html) {
    return <div ref={ref} dangerouslySetInnerHTML={{ __html: props.text }} />;
  } else {
    return <div ref={ref}>{props.text}</div>;
  }
};

KatexRender.defaultProps = {
  html: false
};

export default KatexRender;
