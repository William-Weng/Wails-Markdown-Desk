import exportCSS from '../less/export.less?inline'
import { renderMarkdown } from './markdown'

function escapeHTML(value: string): string {
  return value.replace(/[&<>"']/g, (character) => {
    switch (character) {
      case '&':
        return '&amp;'
      case '<':
        return '&lt;'
      case '>':
        return '&gt;'
      case '"':
        return '&quot;'
      default:
        return '&#39;'
    }
  })
}

export async function buildExportHTML(
  markdown: string,
  documentPath: string,
  title: string,
): Promise<string> {
  const body = await renderMarkdown(markdown, documentPath)

  return `<!doctype html>
<html lang="zh-Hant">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${escapeHTML(title)}</title>
  <style>
${exportCSS}
  </style>
</head>
<body>
  <main class="markdown-body">
${body}
  </main>
</body>
</html>`
}
