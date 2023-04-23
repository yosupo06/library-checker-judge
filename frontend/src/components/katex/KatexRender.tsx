import renderMathInElement from "katex/contrib/auto-render";
import React, { useContext, useEffect, useRef } from "react";
import { LangContext } from "../../contexts/LangContext";

interface Props {
  text: string;
}

const KatexRender: React.FC<Props> = (props) => {
  const lang = useContext(LangContext);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current != null) {
      renderMathInElement(ref.current, {
        delimiters: [
          { left: "$$", right: "$$", display: true },
          { left: "\\[", right: "\\]", display: true },
          { left: "$", right: "$", display: false },
          { left: "\\(", right: "\\)", display: false },
        ],
        ignoredTags: ["script", "noscript", "style"],
      });
      const elems = Array.from(
        ref.current?.getElementsByClassName("lang-ja")
      ).concat(Array.from(ref.current?.getElementsByClassName("lang-en")));
      elems.forEach((e) => {
        if (e.classList.contains(`lang-${lang?.state.lang}`)) {
          e.removeAttribute("hidden");
        } else {
          e.setAttribute("hidden", "true");
        }
      });
    }
  });

  return <div ref={ref} dangerouslySetInnerHTML={{ __html: props.text }} />;
};

export default KatexRender;
