package configure

const ImageWebRoute = "/document-images" // 文件圖片的 HTTP 路由

const DefaultFileName = "Untitled"       // 是尚未儲存文件時使用的預設檔名（不含副檔名）
const MarkdownFilter = "*.md;*.markdown" // 使用者只能看到（或優先看到）副檔名為 .md 或 .markdown 的檔案
const HtmlFilter = "*.html;*.htm"        // 使用者只能看到（或優先看到）副檔名為 .html 或 .htm 的檔案
const MaxDocumentSize = 20 * 1024 * 1024 // 開檔限制為 20 MiB
const MaxImageSize = 10 * 1024 * 1024    // 最大圖片限制為 10 MiB
const EmitKey = "markdown-file-dropped"  // 拖拽markdown事件的Key值
