import { HighlightStyle } from "@codemirror/language";
import { tags } from "@lezer/highlight";

/**
 * Markdown 相關的語法高亮樣式
 */
export const markdownLinkHighlight = HighlightStyle.define([
  {
    tag: [tags.url],
    class: "cm-markdown-link",
  },
  {
    tag: [
      tags.processingInstruction,
      tags.contentSeparator,
      tags.labelName,
    ],
    class: "cm-markdown-marker",
  },
]);

/**
 * 預設載入的 Markdown 範例內容
 */
export const initialMarkdown = `# MarkdownDesk

Markdown, preview, export。

## Features

- Wails
- Svelte 5 + TypeScript
- CodeMirror 6
- markdown-it

\`\`\`go
package main

func main() {
    println("Hello MarkdownDesk")
}
\`\`\`
`;
