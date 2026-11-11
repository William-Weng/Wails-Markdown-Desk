import MarkdownIt from 'markdown-it';
import DOMPurify from 'dompurify';
import hljs from 'highlight.js';

// MARK: - Types

type RenderEnvironment = {
  documentPath: string;
  imageRevision: number;
};

// MARK: - Image Revision

let imageRevision = 0;

/**
 * 更新圖片 URL 的版本參數，讓下次渲染重新請求圖片。
 * 保留既有名稱，相容 App.svelte 的呼叫。
 */
export function clearImageCache(): void {
  imageRevision++;
}

// MARK: - Syntax Highlighting

function highlightCode(code: string, language: string): string {
  if (!language || !hljs.getLanguage(language)) {
    // 交由 markdown-it 使用預設的 HTML escaping。
    return '';
  }

  try {
    return hljs.highlight(code, {
      language,
      ignoreIllegals: true,
    }).value;
  } catch (error) {
    console.warn('程式碼語法高亮失敗：', error);

    return '';
  }
}

// MARK: - Markdown Parser

const parser = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: false,
  breaks: false,
  highlight: highlightCode,
});

// MARK: - Image Rendering

const defaultImageRenderer = parser.renderer.rules.image!;

parser.renderer.rules.image = (
  tokens,
  index,
  options,
  env,
  renderer,
): string => {
  const token = tokens[index];

  if (!token) return '';

  const imageSource = String(token.attrGet('src') ?? '');
  const environment = env as RenderEnvironment;

  token.attrSet('decoding', 'async');

  // 網路圖片直接交給 WebView 載入。
  if (/^https?:\/\//i.test(imageSource)) {
    token.attrSet('referrerpolicy', 'no-referrer');

    return defaultImageRenderer(
      tokens,
      index,
      options,
      env,
      renderer,
    );
  }

  // 尚未開啟或儲存文件，無法解析本機圖片的相對路徑。
  if (!environment.documentPath) {
    const alt = parser.utils.escapeHtml(
      token.content || imageSource || '圖片',
    );

    return (
      '<span class="image-placeholder">' +
      `[${alt}：請先開啟或儲存 Markdown 文件]` +
      '</span>'
    );
  }

  // 透過後端端點載入文件引用的本機圖片。
  const query = new URLSearchParams({
    document: environment.documentPath,
    src: imageSource,
    v: String(environment.imageRevision),
  });

  token.attrSet('src', `/document-images?${query.toString()}`);

  return defaultImageRenderer(
    tokens,
    index,
    options,
    env,
    renderer,
  );
};

// MARK: - Public API

/**
 * 將 Markdown 轉為可供預覽使用的 HTML。
 * 保留 Promise 介面，相容既有的 await 呼叫。
 */
export async function renderMarkdown(
  source: string,
  documentPath: string,
): Promise<string> {
  const environment: RenderEnvironment = {
    documentPath,
    imageRevision,
  };

  const html = parser.render(source, environment);

  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
  });
}
