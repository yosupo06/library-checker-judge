import "katex/dist/katex.min.css";
import React, { useState } from "react";
import Editor from "@monaco-editor/react";
import { editor } from "monaco-editor";

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
  if (lang.startsWith("markdown")) {
    return "markdown";
  }
  return "plaintext";
};

const SourceEditor: React.FC<Props> = (props) => {
  const minHeight = 100;
  const { value, language, onChange, readOnly, autoHeight } = props;
  const [editorHeight, setEditorHeight] = useState(minHeight);

  const mode = editorMode(language);

  const updateHeight = (editor: editor.IStandaloneCodeEditor) => {
    if (autoHeight) {
      setEditorHeight(Math.max(minHeight, editor.getContentHeight()));
    }
  };

  return (
    <Editor
      value={value}
      language={mode}
      height={autoHeight ? editorHeight : undefined}
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
        minimap: {
          enabled: false,
        },
        scrollbar: {
          alwaysConsumeMouseWheel: false,
        },
      }}
    />
  );
};

export default SourceEditor;
