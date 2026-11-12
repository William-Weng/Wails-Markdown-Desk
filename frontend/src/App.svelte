<script lang="ts">
    /// MARK: - golang功能
    import * as FileService from "../../frontend/bindings/markdowndesk/backend/fileservice";
    import * as Model from "../../frontend/bindings/markdowndesk/backend/model/models";

    /// MARK: - Wails功能
    import { tick, onMount } from "svelte";
    import { Events } from "@wailsio/runtime";

    /// MARK: - 外部套件
    import { basicSetup, EditorView } from "codemirror";
    import { markdown } from "@codemirror/lang-markdown";
    import { defaultHighlightStyle, syntaxHighlighting, } from "@codemirror/language";
    import { Compartment } from "@codemirror/state";

    /// MARK: - ts功能
    import { markdownLinkHighlight, initialMarkdown } from "./lib/setting";
    import { renderMarkdown, clearImageCache } from "./lib/markdown";
    import { buildExportHTML } from "./lib/export-html";
    import { loadConfig, getConfig } from "./lib/config";
    import type { ViewMode, StatusBarState } from "./lib/types";

    /// MARK: - 畫面
    import ViewSwitch from "./svelte/ViewSwitch.svelte";
    import ToolbarActions from "./svelte/ToolbarActions.svelte";
    import ToolbarDocument from "./svelte/ToolbarDocument.svelte";
    import ErrorBanner from "./svelte/ErrorBanner.svelte";
    import ExternalFileBanner from "./svelte/ExternalFileBanner.svelte";
    import StatusBar from "./svelte/StatusBar.svelte";
    import EditorPane from "./svelte/EditorPane.svelte";
    import PreviewPane from "./svelte/PreviewPane.svelte";
    import ExportBanner from "./svelte/ExportBanner.svelte";

    /// MARK: - 常數
    const DEFAULT_EDITOR_FONT_SIZE = 14;
    const MIN_EDITOR_FONT_SIZE = 10;
    const MAX_EDITOR_FONT_SIZE = 32;
    const editorFontCompartment = new Compartment();

    const editorThemeSpec: Parameters<typeof EditorView.theme>[0] = {
      "&": {
        height: "100%",
        backgroundColor: "#102a35",
        color: "#d8e5eb",
      },
      ".cm-scroller": {
        overflow: "auto",
        fontFamily:
          "'SFMono-Regular', Consolas, 'Liberation Mono', monospace",
        lineHeight: "1.7",
      },
      ".cm-content": {
        padding: "16px 0",
        caretColor: "#ffffff",
      },
      ".cm-line": {
        padding: "0 16px",
      },
      ".cm-gutters": {
        backgroundColor: "#0c222c",
        color: "#78909c",
        border: "none",
      },
      ".cm-activeLine": {
        backgroundColor: "#ffffff08",
      },
      ".cm-activeLineGutter": {
        backgroundColor: "#ffffff10",
      },
      ".cm-focused .cm-selectionBackground": {
        backgroundColor: "#35596d",
      },
      ".cm-selectionBackground": {
        backgroundColor: "#294858",
      },
      ".cm-cursor": {
        borderLeftColor: "#ffffff",
      },
      "&.cm-focused": {
        outline: "none",
      },
    };

    /// MARK: - 變數
    let previewVersion = 0;
    let fileOperationVersion = 0;

    /// MARK: - 宣告響應式狀態
    let isBusy = $state(false);
    let isConfirmingOpen = $state(false);
    let isConfirmingReload = $state(false);
    let config = $state<Model.AppConfig | null>(null);

    let cursorLine = $state(1);
    let cursorColumn = $state(1);
    let editorFontSize = $state(DEFAULT_EDITOR_FONT_SIZE);

    let errorMessage = $state("");
    let exportMessage = $state("");
    let previewHtml = $state("");
    let filePath = $state("");
    let externalReadError = $state("");
    let fileName = $state("Untitled.md");
    let source = $state(initialMarkdown);
    let savedContent = $state(initialMarkdown);
    let pendingDropPath = $state<string | null>(null);
    let externalContent = $state<string | null>(null);
    let ignoredExternalContent = $state<string | null>(null);

    let viewMode = $state<ViewMode>("split");

    let editorHost = $state<HTMLDivElement | null>(null);
    let editorView = $state<EditorView | null>(null);

    /// MARK: - 計算衍生狀態
    const isDirty = $derived(source !== savedContent);
    const lineCount = $derived(source.split("\n").length);
    const characterCount = $derived(Array.from(source).length);
    const externalConflictMessage = $derived(isConfirmingReload ? "編輯器內容與磁碟內容不同，要從磁碟重新載入嗎？" : "編輯器內容與磁碟內容不同。");
    const pendingDropMessage = $derived(pendingDropPath ? `目前文件尚未儲存。要開啟 ${pendingDropPath} 嗎？` : "");
    const statusState = $derived<StatusBarState>(isBusy ? "busy" : isDirty ? "dirty" : filePath ? "saved" : "empty" );

    onMount(async () => {
      await loadConfig();
      config = await getConfig();
      console.log(config.imageWebRoute, config.emitKey);
    });

    /// MARK: - 處理副作用
    $effect(() => {
        void updatePreview(source, filePath);
    });

    $effect(() => {

      const host = editorHost;
      const currentConfig = config;

      if (!host || !currentConfig) { return; }

      const editor = editorViewMaker(host);
      editorView = editor;
      editor.focus();

      const unsubscribeFileDrop = Events.On(currentConfig.emitKey, (event: any) => {
        const data = event.data as { path?: unknown } | undefined;
          if (!data || typeof data.path !== "string" || data.path.length === 0) { return; }
          handleDroppedFiles([data.path]);
        }
      );

      let checkingOnFocus = false;
      let focusTimer: ReturnType<typeof setTimeout> | undefined;

      const scheduleFocusCheck = () => {
        if (focusTimer !== undefined) { clearTimeout(focusTimer); }
        focusTimer = setTimeout(() => { focusTimer = undefined; void checkOnFocus(); }, 150);
      };

      const handleVisibilityChange = () => {
        if (document.visibilityState === "visible") { scheduleFocusCheck(); }
      };

      const checkOnFocus = async () => {
        if (checkingOnFocus || isBusy || !document.hasFocus() || document.visibilityState === "hidden") { return; }
        checkingOnFocus = true;

        try {
          await checkExternalChanges();
        } finally {
          checkingOnFocus = false;
        }
      };

      document.addEventListener("error", handlePreviewImageError, true);
      window.addEventListener("focus", scheduleFocusCheck);
      document.addEventListener("visibilitychange", handleVisibilityChange);
      window.addEventListener("keydown", handleKeydown, true);

      scheduleFocusCheck();

      return () => {
        unsubscribeFileDrop();

        if (focusTimer !== undefined) { clearTimeout(focusTimer); }

        document.removeEventListener("error", handlePreviewImageError, true);
        window.removeEventListener("focus", scheduleFocusCheck);
        document.removeEventListener("visibilitychange", handleVisibilityChange);
        window.removeEventListener("keydown", handleKeydown, true);

        fileOperationVersion++;
        editorView = null;
        editor.destroy();
      };
    });

    /// MARK: - 一般函數
    /**
     * 建立 Markdown 編輯器的 EditorView
     *
     * 這個函式負責：
     * 1. 將 EditorView 掛載到指定的 DOM 容器
     * 2. 設定初始 Markdown 內容
     * 3. 組合 CodeMirror 的基礎功能、Markdown 語法、外觀主題與字體設定
     * 4. 透過 updateListener 同步編輯器內容與游標位置到外部狀態
     */
    function editorViewMaker(editorHosteditorHost: HTMLElement): EditorView {

      const editor = new EditorView({
          parent: editorHosteditorHost,
          doc: initialMarkdown,
          extensions: [
              basicSetup,
              markdown(),
              EditorView.lineWrapping,
              syntaxHighlighting(defaultHighlightStyle),
              syntaxHighlighting(markdownLinkHighlight),
              editorFontCompartment.of(createEditorFontTheme(DEFAULT_EDITOR_FONT_SIZE)),
              EditorView.theme(editorThemeSpec, { dark: true }),
              EditorView.updateListener.of((update) => {
                  if (update.docChanged) { source = update.state.doc.toString(); }
                  if (update.docChanged || update.selectionSet) {
                      const position = update.state.selection.main.head;
                      const line = update.state.doc.lineAt(position);
                      cursorLine = line.number;
                      cursorColumn = position - line.from + 1;
                  }
              }),
          ],
      });

      return editor;
    }

    /**
     * 處理使用者拖放檔案後的流程
     *
     * 流程：
     * 1. 將拖放的檔案路徑交給 `_validDroppedFilePath()` 驗證
     * 2. 若驗證失敗、正在忙碌、已有錯誤，或文件尚未儲存而需要確認，函式會回傳 `null`
     * 3. 只有取得有效的 Markdown 檔案路徑時，才呼叫 `openDroppedFile()` 開啟檔案
     *
     * @param files - 拖放事件提供的檔案路徑陣列；每個元素代表一個被拖放進來的檔案
     */
    function handleDroppedFiles(files: string[]) {
        const path = validDroppedFilePath(files);
        if (path === null) { return; }
        void openDroppedFile(path);
    }

    /**
     * 要求開啟新的 Markdown 檔案
     *
     * 流程：
     * 1. 若目前正在處理其他工作，或編輯器尚未初始化，直接忽略這次操作
     * 2. 若目前文件有未儲存的變更，先進入確認流程，避免直接覆蓋使用者內容
     * 3. 沒有未儲存變更時，直接開啟檔案選擇器並載入所選檔案
     */
    function requestOpenFile() {
        if (isBusy || !editorView) return;
        if (isDirty) { isConfirmingOpen = true; return; }
        void openFile();
    }

    /**
     * 使用者確認要開啟新檔案後，關閉確認對話框並執行開檔流程
     *
     * 流程：
     * 1. 若目前正在處理其他工作，或編輯器尚未初始化，直接忽略這次操作
     * 2. 關閉「是否放棄未儲存變更」的確認對話框
     * 3. 呼叫 `openFile()` 開啟檔案選擇器並載入所選檔案
     */
    function confirmOpenFile() {
        if (isBusy || !editorView) return;
        isConfirmingOpen = false;
        void openFile();
    }

    /**
     * 使用者確認要開啟拖放進來的 Markdown 檔案後，執行開檔流程
     *
     * 前提：
     * - `pendingDropPath` 是在使用者拖放檔案、但目前文件尚有未儲存變更時，暫時記錄下來的目標檔案路徑
     * - 使用者確認後，才會真正開啟該檔案
     */
    function confirmDroppedFile() {
        if (!pendingDropPath) return;
        void openDroppedFile(pendingDropPath);
    }

    /**
     * 將磁碟上載入的 Markdown 內容套用到編輯器與應用程式狀態
     *
     * 通常在成功開啟檔案、重新載入檔案，或放棄未儲存變更後呼叫
     *
     * 流程：
     * 1. 清除舊文件所產生的圖片快取
     * 2. 將新內容寫入 CodeMirror 編輯器
     * 3. 更新「已儲存內容」與「目前編輯內容」兩個狀態
     * 4. 重設與外部檔案相關的狀態，例如檔案路徑、錯誤訊息或暫存資料
     * 5. 依新內容更新 Markdown 預覽
     *
     * @param content - 從磁碟讀取到的 Markdown 文字內容
     */
    function applyDiskContent(content: string) {

        clearImageCache();
        replaceEditorContent(content);

        savedContent = content;
        source = content;

        resetExternalState();
        void updatePreview(content, filePath);
    }

    /**
     * 處理 Markdown 預覽區塊中圖片載入失敗的情況
     *
     * 當 `<img>` 無法載入時，將它替換成一個顯示替代文字的 `<span>`，避免瀏覽器顯示破圖，同時保留圖片的來源資訊
     *
     * @param event - 圖片元素觸發的 `error` 事件
     */
    function handlePreviewImageError(event: Event) {

        const image = event.target;

        if (!(image instanceof HTMLImageElement)) { return; }
        if (!image.closest(".markdown-body")) { return; }

        const placeholder = document.createElement("span");

        placeholder.className = "image-placeholder";
        placeholder.textContent = image.alt;
        placeholder.title = image.getAttribute("src") ?? "";

        image.replaceWith(placeholder);
    }

    /**
     * 處理編輯器或應用程式視窗的鍵盤快速鍵
     *
     * 支援的快速鍵：
     * - macOS：Cmd + O 開啟檔案；Cmd + S 儲存；Cmd + Shift + S 另存新檔
     * - Windows / Linux：Ctrl + O 開啟檔案；Ctrl + S 儲存；Ctrl + Shift + S 另存新檔
     *
     * @param event - 鍵盤按下事件
     */
    function handleKeydown(event: KeyboardEvent) {

      if (!event.metaKey && !event.ctrlKey) { return; }
      if (event.altKey) { return; }

      switch (event.key.toLowerCase()) {
          case "o": event.preventDefault(); requestOpenFile(); break;
          case "s": event.preventDefault(); saveFile(event.shiftKey); break;
      }
    }

    /**
     * 重設與外部檔案內容相關的暫存狀態
     *
     * 通常在成功載入磁碟內容、關閉確認流程，或需要回到乾淨狀態時呼叫；這個函式不會修改編輯器內容，也不會影響目前文件的儲存狀態
     */
    function resetExternalState() {
        externalContent = null;
        pendingDropPath = null;
        ignoredExternalContent = null;
        externalReadError = "";
        isConfirmingReload = false;
        isConfirmingOpen = false;
    }

    /**
     * 使用者選擇保留目前編輯器內容時，暫存外部內容並結束確認流程
     *
     * 這裡不會直接套用外部內容，也不會改變目前編輯器的文件；外部內容會被記錄到 `ignoredExternalContent`，方便之後提示或重新套用
     */
    function keepEditorContent() {
        ignoredExternalContent = externalContent;
        externalContent = null;
        isConfirmingReload = false;
    }

    /**
     * 將從磁碟載入的檔案資訊與內容，初始化到應用程式狀態
     *
     * 通常在成功開啟 Markdown 檔案後呼叫
     *
     * 流程：
     * 1. 記錄檔案名稱，供 UI 顯示
     * 2. 記錄檔案完整路徑，供後續儲存、重新載入或匯出使用
     * 3. 將磁碟內容記錄為「已儲存內容」，作為判斷文件是否被修改的基準
     * 4. 將同一份內容設定為目前編輯器來源內容
     *
     * @param file - 從磁碟載入的檔案資訊，包含名稱、完整路徑與文字內容
     */
    function initFileContent(file: Model.DocumentFile) {
      fileName = file.name;
      filePath = file.path;
      savedContent = file.content;
      source = file.content;
    }

    /**
     * 取消「開啟新檔案前，是否放棄未儲存變更」的確認流程
     *
     * 只關閉確認對話框，不會開啟檔案，也不會修改目前編輯器內容
     */
    function cancelOpenFile() {
      isConfirmingOpen = false;
    }

    /**
     * 縮小編輯器字體：將目前字級減少 1px
     *
     * 實際套用範圍由 `setEditorFontSize()` 內的最小值限制處理
     */
    function zoomEditorOut() {
      setEditorFontSize(editorFontSize - 1);
    }

    /**
     * 放大編輯器字體：將目前字級增加 1px
     *
     * 實際套用範圍由 `setEditorFontSize()` 內的最大值限制處理
     */
    function zoomEditorIn() {
      setEditorFontSize(editorFontSize + 1);
    }

    /**
     * 將編輯器字體大小重設為預設值
     */
    function resetEditorZoom() {
      setEditorFontSize(DEFAULT_EDITOR_FONT_SIZE);
    }

    /**
     * 將未知型別的錯誤轉換成可顯示的文字
     *
     * @param error - 任意型別的錯誤值，可能來自 `catch (error)`
     * @returns 若是 `Error`，回傳其 `message`；否則回傳 `String(error)` 的結果
     */
    function errorText(error: unknown): string {
      return error instanceof Error ? error.message : String(error);
    }

    /**
     * 顯示「偵測到外部檔案內容已變更，是否重新載入」的確認對話框
     */
    function showReloadConfirmation() {
      isConfirmingReload = true;
    }

    /**
     * 取消「重新載入外部檔案內容」的確認流程
     *
     * 只關閉確認對話框，不會重新載入檔案，也不會修改目前編輯器內容
     */
    function cancelReload(): void {
      isConfirmingReload = false;
    }

    /**
     * 取消待開啟的拖放檔案路徑。
     *
     * 通常在使用者取消「放棄未儲存變更並開啟拖放檔案」時呼叫。
     */
    function cancelPendingDropPath() {
      pendingDropPath = null;
    }

    /// MARK: - 輔助用函數
    /**
     * 驗證使用者拖放進來的檔案路徑，並決定是否可以開啟該 Markdown 檔案
     *
     * 處理順序：
     * 1. 若目前正在處理其他工作，或已有錯誤訊息，直接忽略這次拖放
     * 2. 拖放檔案數量必須正好是 1 個
     * 3. 檔案副檔名必須是 `.md` 或 `.markdown`
     * 4. 若目前文件尚未儲存（isDirty），先記錄待處理路徑，等待使用者確認
     * 5. 全部通過後，回傳檔案路徑
     *
     * @param files - 拖放事件提供的檔案路徑陣列；函式內不會修改此陣列
     * @returns 成功時回傳唯一的 Markdown 檔案路徑；失敗、被忽略或需要等待確認時回傳 `null`
     */
    function validDroppedFilePath(files: readonly string[]): string | null {

      if (isBusy || errorMessage) { return null; }
      if (files.length !== 1) { errorMessage = "請一次只拖放一個 Markdown 檔案。"; return null; }

      const path = files[0];

      if (!path || !/\.(md|markdown)$/i.test(path)) { errorMessage = "僅支援 .md 與 .markdown 檔案。"; return null; }
      if (isDirty) { pendingDropPath = path; return null; }

      return path;
    }

    /**
     * 建立依指定字級調整 CodeMirror 編輯器字體大小的主題
     *
     * @param size - 編輯器內文的字體大小，單位為 px
     * @returns 可加入 CodeMirror Editor extensions 的 ThemeExtension
     */
    function createEditorFontTheme(size: number) {
        return EditorView.theme({
            ".cm-scroller": { fontSize: `${size}px` },
        });
    }

    /**
     * 設定 CodeMirror 編輯器的字體大小
     *
     * 流程：
     * 1. 將傳入的字級限制在允許範圍內，並四捨五入成整數
     * 2. 若新的字級與目前相同，不做任何更新
     * 3. 更新 `editorFontSize` 狀態
     * 4. 透過 `editorFontCompartment` 重新套用字體主題
     * 5. 要求 CodeMirror 重新測量版面，確保行高與捲動位置正確
     *
     * @param size - 使用者設定的字體大小，單位為 px；可為小數或超出範圍的值
     */
    function setEditorFontSize(size: number) {

        const nextSize = Math.min(MAX_EDITOR_FONT_SIZE, Math.max(MIN_EDITOR_FONT_SIZE, Math.round(size)));

        if (nextSize === editorFontSize) { return; }
        editorFontSize = nextSize;

        const editor = editorView;
        if (!editor) { return; }

        editor.dispatch({
            effects: editorFontCompartment.reconfigure(createEditorFontTheme(nextSize))
        });

        editor.requestMeasure();
    }

    /**
     * 將 CodeMirror 編輯器的內容整體替換為指定文字
     *
     * 流程：
     * 1. 確認 EditorView 已初始化
     * 2. 建立一個涵蓋整份文件的變更範圍
     * 3. 將原本內容替換成新的 `content`
     * 4. 將游標移到文件開頭，並捲動到該位置
     * 5. 將焦點移回編輯器，方便使用者立即開始編輯
     *
     * @param content - 要放入編輯器的新文字內容
     */
    function replaceEditorContent(content: string) {

      if (!editorView) { return; }

        editorView.dispatch({
            changes: { from: 0, to: editorView.state.doc.length, insert: content },
            selection: { anchor: 0 },
            scrollIntoView: true,
        });

        editorView.focus();
    }

    /// MARK: - 非同步函數
    /**
     * 開啟使用者拖放進來的 Markdown 檔案
     *
     * 流程：
     * 1. 確認目前沒有其他檔案操作正在進行，且編輯器已初始化
     * 2. 進入忙碌狀態，避免使用者在此期間重複觸發檔案操作
     * 3. 暫時將編輯器設為不可編輯，避免載入過程中使用者修改內容
     * 4. 透過 FileService 讀取檔案內容
     * 5. 清除舊圖片快取、替換編輯器內容、初始化檔案資訊與重設外部狀態
     * 6. 更新 Markdown 預覽
     * 7. 無論成功或失敗，最後都恢復編輯器可編輯狀態，並結束忙碌狀態
     *
     * @param path - 使用者拖放進來的 Markdown 檔案完整路徑
     */
    async function openDroppedFile(path: string) {

        const editor = editorView;

        if (isBusy || !editor) return;

        isBusy = true;
        fileOperationVersion++;
        errorMessage = "";
        pendingDropPath = null;

        const previousEditable = editor.contentDOM.contentEditable;
        editor.contentDOM.contentEditable = "false";

        try {
            const file = await FileService.OpenPath(path);
            if (!file || !editorView || !editor) return;

            clearImageCache();
            replaceEditorContent(file.content);
            initFileContent(file);
            resetExternalState();
            void updatePreview(file.content, file.path);
        } catch (error) {
            errorMessage = errorText(error);
        } finally {
            if (editor) { editor.contentDOM.contentEditable = previousEditable; }
            isBusy = false;
        }
    }

    /**
     * 開啟檔案選擇器，讓使用者選擇並載入一個 Markdown 檔案
     *
     * 流程：
     * 1. 確認目前沒有其他檔案操作正在進行，且編輯器已初始化
     * 2. 進入忙碌狀態，避免使用者在此期間重複觸發檔案操作
     * 3. 暫時將編輯器設為不可編輯，避免載入過程中使用者修改內容
     * 4. 透過 FileService 開啟檔案選擇器並讀取所選檔案
     * 5. 清除舊圖片快取、替換編輯器內容、初始化檔案資訊，並清除待處理的拖放路徑
     * 6. 無論成功或失敗，最後都恢復編輯器可編輯狀態，並結束忙碌狀態
     */
    async function openFile() {

        if (isBusy || !editorView) return;

        isBusy = true;
        errorMessage = "";

        const wasEditable = editorView.contentDOM.contentEditable;
        editorView.contentDOM.contentEditable = "false";

        try {
            const file = await FileService.Open();
            if (!file) return;

            clearImageCache();
            replaceEditorContent(file.content);
            initFileContent(file)
            pendingDropPath = null;
        } catch (error) {
            errorMessage = errorText(error);
        } finally {
            if (editorView) { editorView.contentDOM.contentEditable = wasEditable; }
            isBusy = false;
        }
    }

    /**
     * 儲存目前編輯器中的 Markdown 內容
     *
     * 流程：
     * 1. 若目前正在處理其他檔案操作，直接忽略這次儲存請求
     * 2. 進入忙碌狀態，避免使用者在此期間重複觸發儲存
     * 3. 取得目前編輯器內容的快照，避免非同步期間內容改變造成不一致
     * 4. 透過 FileService 儲存內容；`saveAs` 為 true 時另存新檔
     * 5. 儲存成功後，更新檔案名稱、檔案路徑與「已儲存內容」狀態
     * 6. 依儲存後的內容更新 Markdown 預覽
     * 7. 無論成功或失敗，最後都結束忙碌狀態
     *
     * @param saveAs - 是否以「另存新檔」方式儲存；預設為 false，代表覆寫原檔案
     */
    async function saveFile(saveAs = false) {

        if (isBusy) { return; }
        isBusy = true;
        errorMessage = "";

        const snapshot = source;

        try {
            const file = await FileService.Save(snapshot, saveAs);

            if (!file) return;
            clearImageCache();

            fileName = file.name;
            filePath = file.path;
            savedContent = snapshot;

            void updatePreview(snapshot, filePath);
        } catch (error) {
            errorMessage = errorText(error);
        } finally {
            isBusy = false;
        }
    }

    /**
     * 依目前的 Markdown 原始內容與文件路徑，更新右側預覽 HTML。
     *
     * 為了避免使用者快速輸入時，較舊的渲染結果覆蓋較新的結果，
     * 每次開始渲染前都會遞增 `previewVersion`。
     * 只有當渲染完成時版本仍是最新的，才會更新 `previewHtml`。
     *
     * @param markdownSource - 要渲染成 HTML 的 Markdown 原始文字。
     * @param documentPath - 目前文件在磁碟上的路徑；可用於解析相對路徑圖片或資源。
     */
    async function updatePreview(markdownSource: string, documentPath: string) {

        const version = ++previewVersion;

        try {
            const html = await renderMarkdown(markdownSource, documentPath);
            if (version !== previewVersion) { return; }
            previewHtml = html;
        } catch (error) {
            if (version !== previewVersion) { return; }
            console.error("Markdown preview failed:", error);
        }
    }

    /**
     * 檢查目前開啟的檔案在磁碟上是否已被其他程式修改
     *
     * 流程：
     * 1. 若正在處理檔案操作、尚未開啟檔案，或編輯器尚未初始化，直接結束
     * 2. 記錄本次檢查時的檔案路徑、已儲存內容與檔案操作版本
     * 3. 讀取磁碟上目前的檔案內容
     * 4. 確認等待期間沒有新的檔案操作、檔案路徑或基準內容改變
     * 5. 依磁碟內容與目前編輯器內容的關係，決定是否提示重新載入
     *
     * 狀態判斷：
     * - 磁碟內容與已儲存內容相同：沒有外部變更，清除相關狀態
     * - 磁碟內容與目前編輯器內容相同：使用者已手動同步內容，更新基準內容
     * - 磁碟內容與先前忽略的內容相同：維持原本「已忽略外部變更」的狀態
     * - 其他情況：顯示「外部內容已變更」的確認提示
     *
     * @returns 無回傳值；結果會寫入 `externalContent`、`externalReadError` 等狀態
     */
    async function checkExternalChanges() {

        if (isBusy || !filePath || !editorView) { return; }

        const checkedPath = filePath;
        const checkedBaseline = savedContent;
        const checkedVersion = fileOperationVersion;

        try {
            const file = await FileService.ReadCurrent(checkedPath);

            if (isBusy || checkedVersion !== fileOperationVersion || checkedPath !== filePath || checkedBaseline !== savedContent) { return; }
            if (!file) { externalReadError = "無法讀取目前檔案。"; return; }

            const diskContent = file.content;

            if (diskContent === savedContent) { externalContent = null; ignoredExternalContent = null; return; }
            if (diskContent === source) { savedContent = diskContent; resetExternalState(); return; }
            if (diskContent !== ignoredExternalContent) { externalContent = diskContent; }
        } catch (error) {
            if (!isBusy && checkedVersion === fileOperationVersion && checkedPath === filePath) { externalReadError = errorText(error); }
        }
    }

    /**
     * 放棄目前編輯器中的變更，重新從磁碟載入目前檔案內容。
     *
     * 流程：
     * 1. 確認目前沒有其他檔案操作正在進行、已開啟檔案，且編輯器已初始化
     * 2. 進入忙碌狀態，避免使用者在此期間重複觸發檔案操作
     * 3. 關閉「重新載入外部變更」確認對話框，並清除先前的讀取錯誤
     * 4. 記錄本次操作使用的檔案路徑與版本，避免較舊的非同步結果覆蓋較新的狀態
     * 5. 暫時停用編輯，避免重新載入期間使用者修改內容
     * 6. 讀取磁碟上的最新內容，並套用到編輯器與相關狀態
     * 7. 無論成功或失敗，最後都恢復編輯器可編輯狀態，並結束忙碌狀態
     */
    async function reloadFromDisk() {

        const editor = editorView;
        if (isBusy || !filePath || !editor) return;

        isBusy = true;
        isConfirmingReload = false;
        externalReadError = "";

        const version = fileOperationVersion;
        const path = filePath;
        const previousEditable = editor.contentDOM.contentEditable;

        editor.contentDOM.contentEditable = "false";

        try {
            const file = await FileService.ReadCurrent(path);

            if (version !== fileOperationVersion || path !== filePath || !editorView || !editor) { return; }
            if (!file) { throw new Error("無法讀取檔案。"); }

            applyDiskContent(file.content);
            pendingDropPath = null;
        } catch (error) {
            externalReadError = errorText(error);
        } finally {
            if (editor) { editor.contentDOM.contentEditable = previousEditable; }
            isBusy = false;
        }
    }

    /**
     * 將目前 Markdown 內容匯出成 HTML 檔案
     *
     * 流程：
     * 1. 若目前正在處理其他檔案操作，直接忽略這次匯出請求
     * 2. 進入忙碌狀態，避免使用者在此期間重複觸發匯出
     * 3. 遞增檔案操作版本，讓較早開始、尚未完成的非同步操作可以判斷自己是否已過期
     * 4. 清除前一次操作留下的錯誤與匯出訊息
     * 5. 取得目前編輯器內容、檔案路徑與匯出標題的快照
     * 6. 將 Markdown 渲染成完整 HTML，並透過 FileService 選擇輸出位置與寫入檔案
     * 7. 匯出成功時顯示輸出路徑；失敗時顯示錯誤訊息
     * 8. 無論成功、取消或失敗，最後都結束忙碌狀態
     *
     * @returns 無回傳值；結果會寫入 `exportMessage` 或 `errorMessage`
     */
    async function exportHTML() {

        if (isBusy) { return; }

        isBusy = true;
        fileOperationVersion++;
        errorMessage = "";
        exportMessage = "";

        const snapshot = source;
        const documentPath = filePath;
        const title = fileName.replace(/\.markdown|\.md$/i, "");

        try {
            const html = await buildExportHTML(snapshot, documentPath, title);
            const outputPath = await FileService.ExportHTML(html, documentPath);

            if (!outputPath) { return; }
            exportMessage = `HTML 已匯出：${outputPath}`;
        } catch (error) {
            errorMessage = errorText(error);
        } finally {
            isBusy = false;
        }
    }

    /**
     * 切換檢視模式後，要求 CodeMirror 重新測量編輯器版面
     *
     * 流程：
     * 1. 等待 Svelte 完成本次狀態更新對應的 DOM 更新
     * 2. 若新的模式不是 HTML 預覽模式，表示編輯器會重新顯示
     * 3. 要求 CodeMirror 重新測量版面，確保行高、捲動範圍與可視區域正確
     *
     * @param mode - 要切換到的檢視模式；例如 `"editor"`、`"split"` 或 `"html"`
     */
    async function setViewMode(mode: ViewMode) {
        await tick();
        if (mode !== "html") { editorView?.requestMeasure(); }
    }
</script>

<main class="app" data-file-drop-target>
    <header class="toolbar">
        <ToolbarDocument title="Markdown Desk" {fileName} {filePath} {isDirty} />
        <div class="toolbar-right">
            <ViewSwitch bind:value={viewMode} disabled={isBusy} onchange={setViewMode}/>
            <ToolbarActions disabled={isBusy} onopen={requestOpenFile} onsave={() => saveFile()} onsaveas={() => saveFile(true)} onexport={exportHTML}/>
        </div>
    </header>

    <ErrorBanner message={errorMessage} ondismiss={() => (errorMessage = "")} />
    <ExternalFileBanner variant="error" message={externalReadError} />

    {#if externalContent !== null}
        <ExternalFileBanner variant="conflict" message={externalConflictMessage} disabled={isBusy} confirmLabel={isConfirmingReload ? "重新載入" : "檢視變更"} cancelLabel={isConfirmingReload ? "取消" : "保留編輯器內容"} onconfirm={isConfirmingReload ? reloadFromDisk : showReloadConfirmation} oncancel={isConfirmingReload ? cancelReload : keepEditorContent}/>
    {/if}

    {#if isConfirmingOpen}
        <ExternalFileBanner variant="conflict" message="目前文件尚未儲存。繼續開啟其他檔案後，未儲存的變更將會遺失。" disabled={isBusy} confirmLabel="繼續開啟" cancelLabel="取消" onconfirm={confirmOpenFile} oncancel={cancelOpenFile}/>
    {/if}

    {#if pendingDropPath !== null}
        <ExternalFileBanner variant="conflict" message={pendingDropMessage} disabled={isBusy} confirmLabel="開啟" cancelLabel="取消" onconfirm={confirmDroppedFile} oncancel={cancelPendingDropPath}/>
    {/if}

    <div class="workspace" class:mode-code={viewMode === "code"} class:mode-html={viewMode === "html"} class:mode-split={viewMode === "split"}>
        <EditorPane fontSize={editorFontSize} minFontSize={MIN_EDITOR_FONT_SIZE} maxFontSize={MAX_EDITOR_FONT_SIZE} onzoomout={zoomEditorOut} onzoomin={zoomEditorIn} onzoomreset={resetEditorZoom}>
            <div class="editor-host" bind:this={editorHost}></div>
        </EditorPane>
        <PreviewPane html={previewHtml} />
    </div>

    <ExportBanner message={exportMessage} ondismiss={() => (exportMessage = "")}/>
    <StatusBar {lineCount} {characterCount} {cursorLine} {cursorColumn} {filePath} state={statusState}/>
</main>
