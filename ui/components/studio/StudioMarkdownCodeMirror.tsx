"use client";

import { useMemo } from "react";
import { indentWithTab } from "@codemirror/commands";
import { markdown } from "@codemirror/lang-markdown";
import { EditorView, keymap, type ViewUpdate } from "@codemirror/view";
import CodeMirror from "@uiw/react-codemirror";

import styles from "./Studio.module.css";

type StudioMarkdownCodeMirrorProps = {
  value: string;
  uploading?: boolean;
  onChange: (value: string) => void;
  onFiles: (files: File[]) => void;
  onSelectionChange: (selection: { from: number; to: number }) => void;
};

export function StudioMarkdownCodeMirror({
  value,
  uploading = false,
  onChange,
  onFiles,
  onSelectionChange
}: StudioMarkdownCodeMirrorProps) {
  const extensions = useMemo(
    () => [
      markdown(),
      EditorView.lineWrapping,
      keymap.of([indentWithTab]),
      EditorView.domEventHandlers({
        dragover(event) {
          if (event.dataTransfer?.types.includes("Files")) {
            event.preventDefault();
            return true;
          }
          return false;
        },
        drop(event) {
          const files = Array.from(event.dataTransfer?.files ?? []);
          if (files.length === 0) return false;
          event.preventDefault();
          onFiles(files);
          return true;
        }
      })
    ],
    [onFiles]
  );

  function onUpdate(update: ViewUpdate) {
    if (!update.selectionSet && !update.docChanged) return;
    const selection = update.state.selection.main;
    onSelectionChange({ from: selection.from, to: selection.to });
  }

  return (
    <CodeMirror
      aria-label="Markdown 正文"
      basicSetup={{
        autocompletion: true,
        bracketMatching: true,
        closeBrackets: true,
        foldGutter: false,
        highlightActiveLine: false,
        highlightSelectionMatches: true,
        lineNumbers: true,
        searchKeymap: true
      }}
      className={styles.contentEditorCodeMirror}
      editable={!uploading}
      extensions={extensions}
      height="100%"
      value={value}
      onChange={onChange}
      onUpdate={onUpdate}
    />
  );
}
