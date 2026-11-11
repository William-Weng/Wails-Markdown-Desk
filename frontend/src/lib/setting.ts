import { HighlightStyle } from "@codemirror/language";
import { tags } from "@lezer/highlight";

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
