"use client";

import { useEffect, useMemo, useRef } from "react";
import { indentWithTab } from "@codemirror/commands";
import { markdown } from "@codemirror/lang-markdown";
import { EditorSelection, type Range } from "@codemirror/state";
import { Decoration, type DecorationSet, EditorView, keymap, ViewPlugin, type ViewUpdate } from "@codemirror/view";
import CodeMirror from "@uiw/react-codemirror";

import styles from "./Studio.module.css";

export type MarkdownEditorMode = "live" | "source";

export type MarkdownScrollTarget = {
  id: number;
  line: number;
};

type StudioMarkdownCodeMirrorProps = {
  mode: MarkdownEditorMode;
  scrollTarget?: MarkdownScrollTarget | null;
  value: string;
  uploading?: boolean;
  onChange: (value: string) => void;
  onFiles: (files: File[]) => void;
  onSelectionChange: (selection: { from: number; to: number }) => void;
};

export function StudioMarkdownCodeMirror({
  mode,
  scrollTarget,
  value,
  uploading = false,
  onChange,
  onFiles,
  onSelectionChange
}: StudioMarkdownCodeMirrorProps) {
  const viewRef = useRef<EditorView | null>(null);

  const extensions = useMemo(
    () => [
      markdown(),
      EditorView.lineWrapping,
      mode === "live" ? livePreviewExtension() : [],
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
    [mode, onFiles]
  );

  useEffect(() => {
    if (!scrollTarget || !viewRef.current) return;
    const view = viewRef.current;
    const line = view.state.doc.line(Math.max(1, Math.min(scrollTarget.line, view.state.doc.lines)));
    view.dispatch({
      effects: EditorView.scrollIntoView(line.from, { y: "start" }),
      selection: EditorSelection.cursor(line.from)
    });
    view.focus();
  }, [scrollTarget]);

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
        lineNumbers: mode !== "live",
        searchKeymap: true
      }}
      className={`${styles.contentEditorCodeMirror} ${mode === "live" ? styles.contentEditorCodeMirrorLive : styles.contentEditorCodeMirrorSource}`}
      editable={!uploading}
      extensions={extensions}
      height="100%"
      value={value}
      onChange={onChange}
      onCreateEditor={(view) => {
        viewRef.current = view;
      }}
      onUpdate={onUpdate}
    />
  );
}

function livePreviewExtension() {
  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;

      constructor(view: EditorView) {
        this.decorations = buildLivePreviewDecorations(view);
      }

      update(update: ViewUpdate) {
        if (update.docChanged || update.viewportChanged || update.selectionSet) {
          this.decorations = buildLivePreviewDecorations(update.view);
        }
      }
    },
    {
      decorations: (plugin) => plugin.decorations
    }
  );
}

function buildLivePreviewDecorations(view: EditorView) {
  const activeLine = view.state.doc.lineAt(view.state.selection.main.from).number;
  const decorations: Range<Decoration>[] = [];

  for (const { from, to } of view.visibleRanges) {
    let pos = from;
    while (pos <= to) {
      const line = view.state.doc.lineAt(pos);
      const lineClass = livePreviewLineClass(line.text);
      if (lineClass) {
        decorations.push(Decoration.line({ class: lineClass }).range(line.from));
      }
      if (line.number !== activeLine) {
        collectLivePreviewRanges(line.text, line.from, decorations);
      }
      if (line.to >= to) break;
      pos = line.to + 1;
    }
  }

  decorations.sort((left, right) => left.from - right.from || left.to - right.to);
  return Decoration.set(decorations, true);
}

function livePreviewLineClass(text: string) {
  const heading = /^(#{1,4})\s+/.exec(text);
  if (heading) {
    const headingClass = heading[1].length === 1
      ? styles.livePreviewHeading1
      : heading[1].length === 2
        ? styles.livePreviewHeading2
        : heading[1].length === 3
          ? styles.livePreviewHeading3
          : styles.livePreviewHeading4;
    return `${styles.livePreviewLine} ${headingClass}`;
  }
  if (/^\s*([-*+]|\d+\.)\s+/.test(text)) return `${styles.livePreviewLine} ${styles.livePreviewListLine}`;
  if (text.trim() === "") return `${styles.livePreviewLine} ${styles.livePreviewBlankLine}`;
  return styles.livePreviewLine;
}

function collectLivePreviewRanges(text: string, offset: number, decorations: Range<Decoration>[]) {
  const heading = /^(#{1,4})\s+/.exec(text);
  if (heading) {
    decorations.push(Decoration.replace({ inclusive: false }).range(offset, offset + heading[0].length));
  }

  collectRegexRanges(text, offset, /(\*\*|__)/g, decorations);
  collectRegexRanges(text, offset, /`/g, decorations);

  for (const match of text.matchAll(/!\[([^\]]*)\]\(([^)]+)\)/g)) {
    if (match.index === undefined) continue;
    decorations.push(Decoration.replace({ inclusive: false }).range(offset + match.index, offset + match.index + 2));
    decorations.push(Decoration.replace({ inclusive: false }).range(offset + match.index + 2 + match[1].length, offset + match.index + match[0].length));
  }

  for (const match of text.matchAll(/(?<!!)\[([^\]]+)\]\(([^)]+)\)/g)) {
    if (match.index === undefined) continue;
    decorations.push(Decoration.replace({ inclusive: false }).range(offset + match.index, offset + match.index + 1));
    decorations.push(Decoration.replace({ inclusive: false }).range(offset + match.index + 1 + match[1].length, offset + match.index + match[0].length));
  }
}

function collectRegexRanges(text: string, offset: number, regex: RegExp, decorations: Range<Decoration>[]) {
  for (const match of text.matchAll(regex)) {
    if (match.index === undefined) continue;
    decorations.push(Decoration.replace({ inclusive: false }).range(offset + match.index, offset + match.index + match[0].length));
  }
}
