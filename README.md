# [Wails-Markdown-Desk](https://github.com/William-Weng?tab=repositories&q=wails)

![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-20.19.2-339933?logo=nodedotjs&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v3.0.0-DF0000?logo=wails&logoColor=white)
![LICENSE](https://img.shields.io/github/license/William-Weng/Wails-Markdown-Desk?style=flat&label=LICENSE&color=yellow)
![Tag](https://img.shields.io/github/v/tag/William-Weng/Wails-Markdown-Desk?style=flat&label=Tag)
![Stars](https://img.shields.io/github/stars/William-Weng/Wails-Markdown-Desk?style=flat&label=Stars)

使用 Wails v3、Svelte、TypeScript 與 CodeMirror 6 製作的桌面 Markdown 編輯器。

採用「Markdown 原始碼＋HTML 即時預覽」的雙欄設計，不使用 WYSIWYG 編輯模式。保留 Markdown 原始文字的可控性，同時提供檔案讀寫、圖片預覽、外部變更偵測與 HTML 匯出功能。

https://github.com/user-attachments/assets/dec57cf6-1e0d-4a6f-a4b1-5352492b20d5

## 功能

### Markdown 編輯

- 使用 CodeMirror 6 編輯 Markdown 原始碼。
- 即時更新 HTML 預覽。
- 支援標題、清單、引用、表格、連結與程式碼區塊。
- 以黃色標示 Markdown 語法符號與連結，提高深色編輯區的可讀性。
- 支援搜尋、復原與重做。
- 顯示行數、字元數及游標位置。

### 顯示模式

提供三種顯示方式：

| 模式 | 說明 |
| --- | --- |
| Code | 只顯示 Markdown 原始碼 |
| HTML | 只顯示渲染後的文件 |
| Code + HTML | 左側編輯、右側預覽 |

HTML 模式指渲染後的預覽，不是 HTML 原始碼檢視。

### 編輯器字體大小

- 使用 `A−`、`A+` 調整左側編輯區的文字大小。
- 支援重設為預設大小。
- 調整範圍為 10–32px，預設為 14px。
- 不影響右側預覽及匯出 HTML 的字體大小。
- 不會因調整字體而將文件標記為未儲存。
- 目前不跨程式啟動保存字體設定。

### 檔案操作

- 開啟 `.md`、`.markdown` 文件。
- 儲存與另存新檔。
- 使用原生檔案對話框選擇位置。
- 顯示目前檔名與未儲存標記。
- 支援 UTF-8 文字文件。

### 外部文件變更

切回 MarkdownDesk 時，檢查目前文件是否被其他程式修改，不持續輪詢磁碟。

| 狀況 | 處理方式 |
| --- | --- |
| Editor 沒有未儲存修改 | 自動載入磁碟最新內容 |
| Editor 也有未儲存修改 | 顯示衝突提示，不直接覆蓋 |
| 選擇重新載入 | 二次確認後，以磁碟版本取代 Editor 內容 |
| 選擇保留編輯 | 保留 Editor 內容，不立即寫回磁碟 |
| 文件無法讀取 | 顯示錯誤並保留 Editor 內容 |

儲存回目前文件前，也會檢查是否有外部修改。

### 圖片預覽

- 支援本機相對路徑圖片。
- 支援 HTTP／HTTPS 網路圖片。
- 本機支援 PNG、JPEG、GIF、WebP。
- 預覽圖片透過 Wails service 路由載入，不使用 Base64。
- 圖片依預覽區寬度縮放，不超出容器。
- 圖片載入失敗時顯示提示。

本機圖片以目前 Markdown 文件所在資料夾為基準：

```text
my-notes/
├── Demo.md
└── assets/
    └── example.png
```

### HTML 匯出

- 匯出目前 Editor 的內容，不必先儲存 Markdown。
- 產生包含樣式的完整 HTML 文件。
- 文章不設定固定最大寬度，使用視窗可用寬度。
- 本機圖片引用原始位置，不複製、不轉成 Base64。
- 根據 HTML 的輸出位置，重新計算圖片相對路徑。
- 網路圖片保留原始 URL。
- 包含 A4 列印樣式。

匯出 HTML 不會將 Markdown 文件標記為已儲存。

## 技術組成

| 項目 | 技術 |
| --- | --- |
| 桌面框架 | Wails v3 |
| 後端 | Go |
| 前端 | Svelte、TypeScript |
| 編輯器 | CodeMirror 6 |
| Markdown 解析 | markdown-it |
| HTML 清理 | DOMPurify |
| 樣式 | Less |
| 前端建置 | Vite |
| 匯出 HTML 處理 | golang.org/x/net/html |

## 開發環境

請先準備：

- 符合專案 `go.mod` 要求的 Go 版本。
- Node.js 與 npm。
- Wails v3 CLI。
- 作業系統所需的 Wails 開發依賴。

本機圖片處理使用 `os.OpenRoot`，因此 Go 版本至少需要 1.24；實際最低需求仍以專案與所用 Wails 版本為準。

檢查開發環境：

```bash
wails3 doctor
```

本專案使用 Wails v3，請使用 `wails3`，不是 Wails v2 的 `wails` 指令。

## 快捷鍵

| 操作 | macOS | Windows / Linux |
| --- | --- | --- |
| 開啟文件 | Command + O | Ctrl + O |
| 儲存 | Command + S | Ctrl + S |
| 另存新檔 | Command + Shift + S | Ctrl + Shift + S |

## HTML 匯出與圖片路徑

假設來源文件如下：

```text
my-notes/
├── Demo.md
└── assets/
    └── example.png
```

如果匯出到同一個資料夾：

```text
my-notes/
├── Demo.md
├── Demo.html
└── assets/
    └── example.png
```

HTML 會引用：

```html
<img src="assets/example.png" alt="範例圖片">
```

如果匯出到子資料夾：

```text
my-notes/
├── Demo.md
├── exports/
│   └── Demo.html
└── assets/
    └── example.png
```

HTML 會引用：

```html
<img src="../assets/example.png" alt="範例圖片">
```

匯出的 HTML 可在不啟動 MarkdownDesk 的情況下開啟，但圖片仍依賴原始檔案與相對目錄結構。

移動或分享文件時，請一併保留圖片與 HTML 的相對位置。網路圖片需要可用的網路連線。

## 安全與資料處理

- Markdown 原始 HTML 解析目前關閉。
- 預覽內容經 DOMPurify 清理。
- 本機圖片存取限制在目前 Markdown 文件所在資料夾內。
- 不支援以任意絕對路徑或 `file://` 存取本機圖片。
- 使用 `os.OpenRoot` 限制檔案存取範圍。
- 匯出時不把 App 專用的圖片路由寫入最終 HTML。

## [建置指令整理](https://v3.wails.io/zh-tw/guides/build/building/)

| 目的 | 指令 |
| --- | --- |
| 建立新專案 | `wails3 init -n <專案名稱> -t <前端框架>` |
| 產生 bindings | `wails3 generate bindings` |
| 更新 Windows 與 macOS 專用圖示檔案 | `wails3 generate icons -input build/appicon.png -windowsfilename build/windows/icon.ico -macfilename build/darwin/icons.icns` |
| 更新建置資源 | `wails3 update build-assets -config build/config.yml -dir build` |
| 拉取（下載）用於跨平台交叉編譯的 Docker 映像檔 | `wails3 task setup:docker` |
| 打包 macOS arm64 | `wails3 package GOOS=darwin GOARCH=arm64` |
| 建置 Windows x64 | `wails3 build GOOS=windows GOARCH=amd64` |
| 打包 Linux arm64 | `wails3 build GOOS=linux GOARCH=arm64` |
