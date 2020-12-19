import renderMathInElement from "katex/dist/contrib/auto-render";
import "katex/dist/katex.min.css";
import React, { useContext, useState } from "react";
import { LangContext } from "../contexts/LangContext";
import { ControlledEditor } from "@monaco-editor/react";

interface Props {
  value: string;
  language?: string;
  onChange?: (value: string) => void;
  readOnly: boolean;
  autoHeight: boolean;
}

const editorMode = (lang?: string) => {
  if (!lang) {
    return "plaintext";
  }
  if (lang.startsWith("cpp")) {
    return "cpp";
  }
  if (lang.startsWith("java")) {
    return "java";
  }
  if (lang.startsWith("py")) {
    return "python";
  }
  if (lang.startsWith("rust")) {
    return "rust";
  }
  if (lang.startsWith("d")) {
    return "plaintext";
  }
  if (lang.startsWith("haskell")) {
    return "plaintext";
  }
  if (lang.startsWith("csharp")) {
    return "csharp";
  }
  if (lang.startsWith("go")) {
    return "go";
  }
  if (lang.startsWith("lisp")) {
    return "plaintext";
  }
  return "plaintext";
};

const Editor: React.FC<Props> = props => {
  const { value, language, onChange, readOnly, autoHeight } = props;
  const [editorHeight, setEditorHeight] = useState(100);

  const mode = editorMode(language);

  return (
    <ControlledEditor
      value={value}
      language={mode}
      height={autoHeight ? editorHeight : undefined}
      onChange={(_, e) => {
        if (e !== undefined && onChange) onChange(e);
      }}
      editorDidMount={(_, editor) => {
        if (autoHeight) setEditorHeight(editor.getContentHeight() + 18);
      }}
      options={{
        readOnly: readOnly,
        scrollBeyondLastColumn: 0,
        scrollBeyondLastLine: false,
        minimap: {
          enabled: false
        },
        scrollbar: {
          alwaysConsumeMouseWheel: false
        }
      }}
    />
  );
};

export default Editor;
