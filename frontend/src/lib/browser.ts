import { Browser } from "@wailsio/runtime";

/**
 * 判斷是否為外部 http/https 連結
 *
 * - 支援完整網址，例如 https://example.com
 * - 也支援相對路徑，會以目前頁面網址為 base 解析
 * - 若解析失敗，視為非外部連結，回傳 false
 */
function isExternalHttpUrl(href: string): boolean {

  try {
    const url = new URL(href, window.location.href);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

// MARK: - Public API
/**
 * 處理 Markdown 預覽區內所有 <a> 連結的點擊事件
 *
 * 目的：
 * - 外部 http/https 連結 → 阻止在 Wails WebView 內跳轉，改用系統預設瀏覽器開啟
 * - 內部 anchor（如 #section）或相對路徑 → 不攔截，維持原本行為
 */
export function handleClick(event: MouseEvent) {

  const target = event.target as HTMLElement | null;
  const anchor = target?.closest("a");

  if (!anchor) { return; }

  const href = anchor.getAttribute("href");
  if (!href || !isExternalHttpUrl(href)) return;

  event.preventDefault();
  void Browser.OpenURL(href).catch((error) => { console.error("Failed to open external URL:", error); });
}
