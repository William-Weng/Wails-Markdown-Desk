import MarkdownIt from 'markdown-it';
import DOMPurify from 'dompurify';
import hljs from 'highlight.js';
import { getConfig } from "./config";

/**
 * 渲染 Markdown 時需要的環境資訊
 */
type RenderEnvironment = {
  documentPath: string;
  imageRevision: number;
  imageWebRoute: string;
};

/** 圖片版本號，每次 clearImageCache() 時 +1 */
let imageRevision = 0;

/**
 * 建立 MarkdownIt parser
 *
 * - html: false → 不允許 Markdown 裡直接寫原始 HTML（較安全）
 * - linkify: true → 自動把網址文字轉成 <a>
 * - typographer: false → 不做智慧引號等排版替換
 * - breaks: false → 單行換行不會變 <br>
 * - highlight: highlightCode → 程式碼區塊使用 highlight.js
 */
const parser = new MarkdownIt({ html: false, linkify: true, typographer: false, breaks: false, highlight: highlightCode });

/** 保留 MarkdownIt 預設的 <img> renderer，方便之後呼叫 */
const defaultImageRenderer = parser.renderer.rules.image!;

/**
 * 自訂 <img> renderer：
 * - 網路圖片 → 直接交給 WebView 載入
 * - 本機圖片 → 改用後端端點 /document-images 載入
 * - 尚未開啟文件 → 顯示提示文字
 */
parser.renderer.rules.image = (tokens, index, options, env, renderer): string => {

  const token = tokens[index];
  if (!token) return '';

  const imageSource = String(token.attrGet('src') ?? '');
  const environment = env as RenderEnvironment;

  token.attrSet('decoding', 'async');

  // 網路圖片直接交給 WebView 載入。
  if (/^https?:\/\//i.test(imageSource)) {
    token.attrSet('referrerpolicy', 'no-referrer');
    return defaultImageRenderer(tokens, index, options, env, renderer);
  }

  // 尚未開啟或儲存文件，無法解析本機圖片的相對路徑。
  if (!environment.documentPath) {
    const alt = parser.utils.escapeHtml(token.content || imageSource || '圖片');
    return ('<span class="image-placeholder">' + `[${alt}：請先開啟或儲存 Markdown 文件]` + '</span>');
  }

  // 透過後端端點載入文件引用的本機圖片。
  const query = new URLSearchParams({ document: environment.documentPath, src: imageSource, v: String(environment.imageRevision) });
  token.attrSet('src', `${environment.imageWebRoute}/document-images?${query.toString()}`);
  return defaultImageRenderer(tokens, index, options, env, renderer);
};

/**
 * 程式碼語法高亮
 *
 * - 有指定語言且 highlight.js 支援 → 高亮
 * - 沒有語言或不支援 → 回傳空字串（MarkdownIt 會 fallback）
 * - 高亮失敗 → 只警告，不中斷渲染
 */
function highlightCode(code: string, language: string): string {

  if (!language || !hljs.getLanguage(language)) { return ''; }

  try {
    return hljs.highlight(code, { language, ignoreIllegals: true }).value;
  } catch (error) {
    console.warn('程式碼語法高亮失敗：', error);
    return '';
  }
}

// MARK: - Public API
/**
 * 更新圖片 URL 的版本參數，讓下次渲染重新請求圖片。
 * 保留既有名稱，相容 App.svelte 的呼叫。
 */
export function clearImageCache(): void {
  imageRevision++;
}

/**
 * 將 Markdown 轉為可供預覽使用的 HTML
 * 保留 Promise 介面，相容既有的 await 呼叫
 */
export async function renderMarkdown(source: string, documentPath: string): Promise<string> {

  const config = await getConfig();
  const environment: RenderEnvironment = { documentPath, imageRevision, imageWebRoute: config.imageWebRoute };
  const html = parser.render(source, environment);

  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
}
