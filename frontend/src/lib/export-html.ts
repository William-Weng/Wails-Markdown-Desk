import exportCSS from '../less/export.less?inline'
import { renderMarkdown } from './markdown'

/**
 * 將 Markdown 內容組合成可直接匯出／儲存的完整 HTML 文件
 *
 * @param markdown      原始 Markdown 文字
 * @param documentPath  Markdown 文件所在路徑，通常用來解析相對圖片或資源
 * @param title         HTML <title> 與文件標題
 * @returns             完整的 HTML 字串
 */
export async function buildExportHTML(markdown: string, documentPath: string, title: string,): Promise<string> {

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

/**
 * 將字串中的 HTML 特殊字元跳脫，避免破壞 HTML 結構或造成 XSS
 *
 * 主要用在 <title> 這類會被直接插入 HTML 的文字
 */
function escapeHTML(value: string): string {

  return value.replace(/[&<>"']/g, (character) => {
    switch (character) {
      case '&': return '&amp;'
      case '<': return '&lt;'
      case '>': return '&gt;'
      case '"': return '&quot;'
      default: return '&#39;'
    }
  })
}
