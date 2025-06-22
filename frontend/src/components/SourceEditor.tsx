import "katex/dist/katex.min.css";
import React, { useState } from "react";
import Editor from "@monaco-editor/react";
import { editor } from "monaco-editor";
import { Box } from "@mui/material";

interface Props {
  value: string;
  language?: string;
  onChange?: (value: string) => void;
  readOnly: boolean;
  height?: number;
  placeholder?: string;
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
  if (lang.startsWith("markdown")) {
    return "markdown";
  }
  return "plaintext";
};

const MIN_EDITOR_HEIGHT = 100;

const SourceEditor: React.FC<Props> = (props) => {
  const { value, language, onChange, readOnly, height, placeholder } = props;
  const [editorHeight, setEditorHeight] = useState(height ?? MIN_EDITOR_HEIGHT);

  const mode = editorMode(language);

  const updateHeight = (editor: editor.IStandaloneCodeEditor) => {
    if (height === undefined) {
      setEditorHeight(Math.max(MIN_EDITOR_HEIGHT, editor.getContentHeight()));
    }
  };

  return (
    <Box sx={{ width: "auto" }}>
      <Editor
        value={value}
        language={mode}
        height={editorHeight}
        onChange={(src) => {
          if (src !== undefined && onChange) onChange(src);
        }}
        onMount={(editor) => {
          editor.onDidContentSizeChange(() => {
            updateHeight(editor);
          });
          updateHeight(editor);
        }}
        options={{
          readOnly: readOnly,
          scrollBeyondLastColumn: 0,
          scrollBeyondLastLine: false,
          placeholder: placeholder,
          minimap: {
            enabled: false,
          },
          scrollbar: {
            alwaysConsumeMouseWheel: false,
          },
        }}
      />
    </Box>
  );
};

export default SourceEditor;
