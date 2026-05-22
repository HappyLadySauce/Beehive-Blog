import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type React from "react";

export function AnzhiyuMarkdown({ children }: { children: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        h2: ({ children: headingChildren }) => {
          const text = plainText(headingChildren);
          return <h2 id={headingID(text)}>{headingChildren}</h2>;
        },
        h3: ({ children: headingChildren }) => {
          const text = plainText(headingChildren);
          return <h3 id={headingID(text)}>{headingChildren}</h3>;
        }
      }}
    >
      {children}
    </ReactMarkdown>
  );
}

export function extractToc(markdown: string) {
  return markdown
    .split(/\r?\n/)
    .map((line) => /^(##|###)\s+(.+)$/.exec(line.trim()))
    .filter((match): match is RegExpExecArray => Boolean(match))
    .map((match) => ({
      id: headingID(match[2]),
      text: match[2].replace(/[#*_`]/g, "").trim(),
      level: match[1].length
    }));
}

function headingID(text: string) {
  return text
    .replace(/[#*_`]/g, "")
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "");
}

function plainText(value: React.ReactNode): string {
  if (typeof value === "string" || typeof value === "number") return String(value);
  if (Array.isArray(value)) return value.map(plainText).join("");
  return "";
}
