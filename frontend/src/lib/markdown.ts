import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'

const parser = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: false,
  breaks: false,
})

type RenderEnvironment = {
  documentPath: string
  imageRevision: number
}

let imageRevision = 0

// 保留既有函式名稱，App.svelte 不必跟著改。
// 現在的用途是更新圖片 URL，讓重新載入時重新請求圖片。
export function clearImageCache() {
  imageRevision++
}

const defaultImageRenderer = parser.renderer.rules.image!

parser.renderer.rules.image = (
  tokens,
  index,
  options,
  env,
  renderer,
) => {
  const token = tokens[index]
  if (!token) return ''

  const imageSource = String(token.attrGet('src') ?? '')
  const environment = env as RenderEnvironment

  // 網路圖片直接由 WebView 載入。
  if (/^https?:\/\//i.test(imageSource)) {
    token.attrSet('referrerpolicy', 'no-referrer')
    token.attrSet('decoding', 'async')

    return defaultImageRenderer(
      tokens,
      index,
      options,
      env,
      renderer,
    )
  }

  if (!environment.documentPath) {
    const alt = parser.utils.escapeHtml(
      token.content || imageSource || '圖片',
    )

    return (
      '<span class="image-placeholder">' +
      `[${alt}：請先開啟或儲存 Markdown 文件]` +
      '</span>'
    )
  }

  const query = new URLSearchParams({
    document: environment.documentPath,
    src: imageSource,
    v: String(environment.imageRevision),
  })

  token.attrSet('src', `/document-images?${query.toString()}`)
  token.attrSet('decoding', 'async')

  return defaultImageRenderer(
    tokens,
    index,
    options,
    env,
    renderer,
  )
}

// 保留 Promise 介面，相容目前 App.svelte 的 await 呼叫。
export async function renderMarkdown(
  source: string,
  documentPath: string,
): Promise<string> {
  const environment: RenderEnvironment = {
    documentPath,
    imageRevision,
  }

  const html = parser.render(source, environment)

  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
  })
}
