<script lang="ts">
    import { onMount, tick } from "svelte";
    import { basicSetup, EditorView } from "codemirror";
    import { markdown } from "@codemirror/lang-markdown";
    import {
      defaultHighlightStyle,
      HighlightStyle,
      syntaxHighlighting,
    } from '@codemirror/language'

    import { tags } from '@lezer/highlight'
    import { renderMarkdown, clearImageCache } from "./lib/markdown";
    import { buildExportHTML } from './lib/export-html'
    import * as FileService from "../../frontend/bindings/MarkdownDesk/fileservice";

    const initialMarkdown = [
        "# MarkdownDesk",
        "",
        "左側編輯 Markdown，右側即時預覽。",
        "",
        "## 文字格式",
        "",
        "這是 **粗體**、*斜體* 與 `inline code`。",
        "",
        "## 清單",
        "",
        "- Wails",
        "- Svelte + TypeScript",
        "- CodeMirror 6",
        "- markdown-it",
        "",
        "## 表格",
        "",
        "| 功能 | 狀態 |",
        "| --- | --- |",
        "| 雙欄編輯 | 完成 |",
        "| 即時預覽 | 完成 |",
        "| 開檔與存檔 | 下一階段 |",
        "",
        "## 程式碼",
        "",
        "```go",
        "package main",
        "",
        "func main() {",
        '    println("Hello MarkdownDesk")',
        "}",
        "```",
        "",
        "> 預覽區沒有固定的最大文章寬度。",
        "",
        "---",
        "",
        "![範例圖片](./assets/example.png)",
        "",
    ].join("\n");

    let editorHost: HTMLDivElement;
    let source = initialMarkdown;
    let previewHtml = ''
    let lineCount = initialMarkdown.split("\n").length;
    let characterCount = Array.from(initialMarkdown).length;
    let cursorLine = 1;
    let cursorColumn = 1;
    let confirmingReload = false
    let editorView: EditorView | null = null;

    let fileName = "Untitled.md";
    let filePath = "";
    let savedContent = initialMarkdown;
    let isDirty = false;
    let isBusy = false;
    let errorMessage = "";
    let externalContent: string | null = null
    let ignoredExternalContent: string | null = null
    let externalReadError = ''

    // 開檔、儲存時遞增，讓舊的背景讀取結果失效。
    let fileOperationVersion = 0

    $: isDirty = source !== savedContent;

    type ViewMode = "code" | "html" | "split";

    let viewMode: ViewMode = "split";

    const markdownLinkHighlight = HighlightStyle.define([
      {
        tag: [tags.link, tags.url],
        class: 'cm-markdown-link',
      },
      {
        tag: [
          tags.processingInstruction,
          tags.contentSeparator,
        ],
        class: 'cm-markdown-marker',
      },
    ])

    async function setViewMode(mode: ViewMode) {
        if (viewMode === mode) return;

        viewMode = mode;

        // 等 Svelte 套用顯示模式與版面。
        await tick();

        if (mode !== "html") {
            editorView?.requestMeasure();
        }
    }
    let previewVersion = 0

    async function updatePreview(
      markdownSource: string,
      documentPath: string,
    ) {
      const version = ++previewVersion

      try {
        const html = await renderMarkdown(
          markdownSource,
          documentPath,
        )

        // 避免較舊的非同步結果覆蓋最新文件。
        if (version !== previewVersion) return

        previewHtml = html
      } catch (error) {
        if (version !== previewVersion) return

        console.error('Markdown 預覽失敗', error)
      }
    }

    $: void updatePreview(source, filePath)

    function errorText(error: unknown): string {
        return error instanceof Error ? error.message : String(error);
    }

    function replaceEditorContent(content: string) {
        if (!editorView) return;

        editorView.dispatch({
            changes: {
                from: 0,
                to: editorView.state.doc.length,
                insert: content,
            },
            selection: { anchor: 0 },
            scrollIntoView: true,
        });

        editorView.focus();
    }

    async function openFile() {
        if (isBusy || !editorView) return;

        if (
            isDirty &&
            !window.confirm(
                "目前有尚未儲存的修改，確定捨棄修改並開啟其他檔案？",
            )
        ) {
            return;
        }

        isBusy = true;
        errorMessage = "";

        // 對話框與讀檔期間，避免新的編輯被讀回的文件覆蓋。
        const wasEditable = editorView.contentDOM.contentEditable;
        editorView.contentDOM.contentEditable = "false";

        try {
          const file = await FileService.Open()
          if (!file) return

          clearImageCache()

          replaceEditorContent(file.content)

          fileName = file.name
          filePath = file.path
          savedContent = file.content
        } catch (error) {
            errorMessage = errorText(error);
        } finally {
            editorView.contentDOM.contentEditable = wasEditable;
            isBusy = false;
        }
    }

    async function saveFile(saveAs = false) {
        if (isBusy) return;

        isBusy = true;
        errorMessage = "";

        // 儲存的是按下按鈕當下的內容快照。
        const snapshot = source;

        try {
          const file = await FileService.Save(snapshot, saveAs)
          if (!file) return

          clearImageCache()

          fileName = file.name
          filePath = file.path
          savedContent = snapshot

          // 即使路徑沒有改變，也重新載入預覽圖片。
          void updatePreview(source, filePath)
        } catch (error) {
            errorMessage = errorText(error);
        } finally {
            isBusy = false;
        }
    }

    function resetExternalState() {
      externalContent = null
      ignoredExternalContent = null
      externalReadError = ''
      confirmingReload = false
    }

    function applyDiskContent(content: string) {
      clearImageCache()

      replaceEditorContent(content)
      savedContent = content

      resetExternalState()

      // 即使 Markdown 文字相同，也允許更新圖片預覽。
      void updatePreview(content, filePath)
    }

    async function checkExternalChanges() {
      if (isBusy || !filePath || !editorView) return

      const checkedPath = filePath
      const checkedBaseline = savedContent
      const checkedVersion = fileOperationVersion

      try {
        const file = await FileService.ReadCurrent(checkedPath)

        // 等待 Go 回覆期間可能發生開檔、存檔或文件切換。
        if (
          !file ||
          isBusy ||
          checkedVersion !== fileOperationVersion ||
          checkedPath !== filePath ||
          checkedBaseline !== savedContent
        ) {
          return
        }

        externalReadError = ''

        const diskContent = file.content

        // 磁碟與最後載入／儲存版本相同。
        if (diskContent === savedContent) {
          externalContent = null
          ignoredExternalContent = null
          return
        }

        // 外部程式剛好寫入與 Editor 相同的內容。
        if (diskContent === source) {
          savedContent = diskContent
          resetExternalState()
          return
        }

        // 直接比較 source，避免依賴 reactive isDirty 的更新時機。
        if (source === savedContent) {
          applyDiskContent(diskContent)
          return
        }

        // Editor 也有修改：保留內容，提示衝突。
        if (diskContent !== ignoredExternalContent) {
          externalContent = diskContent
        }
      } catch (error) {
        if (
          !isBusy &&
          checkedVersion === fileOperationVersion &&
          checkedPath === filePath
        ) {
          externalReadError = errorText(error)
        }
      }
    }

    async function reloadFromDisk() {
      const editor = editorView

      if (isBusy || !filePath || !editor) return

      isBusy = true
      confirmingReload = false
      externalReadError = ''

      const version = ++fileOperationVersion
      const path = filePath

      // 讀取期間暫停輸入，避免剛輸入的內容被覆蓋。
      const previousEditable = editor.contentDOM.contentEditable
      editor.contentDOM.contentEditable = 'false'

      try {
        const file = await FileService.ReadCurrent(path)

        if (
          version !== fileOperationVersion ||
          path !== filePath ||
          editorView !== editor
        ) {
          return
        }

        if (!file) {
          throw new Error('目前文件已切換，無法重新載入')
        }

        applyDiskContent(file.content)
      } catch (error) {
        externalReadError = errorText(error)
      } finally {
        editor.contentDOM.contentEditable = previousEditable
        isBusy = false
      }
    }

    function keepEditorContent() {
      ignoredExternalContent = externalContent
      externalContent = null
      confirmingReload = false
    }

    let exportMessage = ''

    async function exportHTML() {
      if (isBusy) return

      isBusy = true
      fileOperationVersion++

      errorMessage = ''
      exportMessage = ''

      const snapshot = source
      const documentPath = filePath
      const title = fileName.replace(/\.markdown$|\.md$/i, '')

      try {
        const html = await buildExportHTML(
          snapshot,
          documentPath,
          title,
        )

        const outputPath = await FileService.ExportHTML(
          html,
          documentPath,
        )

        if (!outputPath) return

        exportMessage = `HTML 已匯出：${outputPath}`
      } catch (error) {
        errorMessage = errorText(error)
      } finally {
        isBusy = false
      }
    }

    function handleKeydown(event: KeyboardEvent) {
        if (!event.metaKey && !event.ctrlKey) return;
        if (event.altKey) return;

        switch (event.key.toLowerCase()) {
            case "o":
                event.preventDefault();
                void openFile();
                break;

            case "s":
                event.preventDefault();
                void saveFile(event.shiftKey);
                break;
        }
    }

    function handlePreviewImageError(event: Event) {
      const image = event.target

      if (
        !(image instanceof HTMLImageElement) ||
        !image.closest('.markdown-body')
      ) {
        return
      }

      const placeholder = document.createElement('span')

      placeholder.className = 'image-placeholder'
      placeholder.textContent =
        `[圖片載入失敗：${image.alt || '未命名圖片'}]`

      placeholder.title = image.getAttribute('src') || ''

      image.replaceWith(placeholder)
    }

    onMount(() => {
        const editor = new EditorView({
            parent: editorHost,
            doc: initialMarkdown,

            extensions: [
                basicSetup,
                markdown(),
                EditorView.lineWrapping,

                // 明確保留其他語法的預設高亮。
                syntaxHighlighting(defaultHighlightStyle),

                // 為連結與網址加上自訂 class。
                syntaxHighlighting(markdownLinkHighlight),

                EditorView.theme(
                    {
                        "&": {
                            height: "100%",
                            backgroundColor: "#102a35",
                            color: "#d8e5eb",
                        },

                        ".cm-scroller": {
                            overflow: "auto",
                            fontFamily:
                                '"SFMono-Regular", Consolas, "Liberation Mono", monospace',
                            fontSize: "14px",
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

                        "&.cm-focused .cm-selectionBackground": {
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
                    },
                    { dark: true },
                ),

                EditorView.updateListener.of((update) => {
                  if (update.docChanged) {
                    source = update.state.doc.toString()
                    lineCount = update.state.doc.lines
                    characterCount = Array.from(source).length
                  }

                    if (update.docChanged || update.selectionSet) {
                        const position = update.state.selection.main.head;
                        const line = update.state.doc.lineAt(position);

                        cursorLine = line.number;
                        cursorColumn = position - line.from + 1;
                    }
                }),
            ],
        });

        editorView = editor;
        editor.focus();

        let disposed = false
        let checkingOnFocus = false
        let focusTimer: ReturnType<typeof setTimeout> | undefined

        document.addEventListener('error', handlePreviewImageError, true)

        function scheduleFocusCheck() {
          if (focusTimer !== undefined) {
            clearTimeout(focusTimer)
          }

          // 稍微延後，等焦點狀態穩定；
          // 同時合併 focus 與 visibilitychange 的連續事件。
          focusTimer = setTimeout(() => {
            focusTimer = undefined
            void checkOnFocus()
          }, 150)
        }

        async function checkOnFocus() {
          if (
            disposed ||
            checkingOnFocus ||
            isBusy ||
            !document.hasFocus() ||
            document.visibilityState === 'hidden'
          ) {
            return
          }

          checkingOnFocus = true

          try {
            await checkExternalChanges()
          } finally {
            checkingOnFocus = false
          }
        }

        function handleVisibilityChange() {
          if (document.visibilityState === 'visible') {
            scheduleFocusCheck()
          }
        }

        window.addEventListener('focus', scheduleFocusCheck)
        document.addEventListener(
          'visibilitychange',
          handleVisibilityChange,
        )

        // 處理 component 建立時視窗已經有焦點的情況。
        scheduleFocusCheck()

        // 使用 capture，先攔截檔案快捷鍵。
        window.addEventListener("keydown", handleKeydown, true);

        return () => {
          disposed = true

          if (focusTimer !== undefined) {
            clearTimeout(focusTimer)
          }

          window.removeEventListener('focus', scheduleFocusCheck)
          document.removeEventListener('error', handlePreviewImageError, true)

          document.removeEventListener(
            'visibilitychange',
            handleVisibilityChange,
          )

          // 讓尚未完成的外部檢查結果失效。
          fileOperationVersion++

          window.removeEventListener('keydown', handleKeydown, true)

          editorView = null
          editor.destroy()
        };
    });
</script>

<main class="app">
    <header class="toolbar">
        <div class="toolbar-document" title={filePath || fileName}>
            <span class="app-title">MarkdownDesk</span>

            <span class="document-name">
                {fileName}{isDirty ? " *" : ""}
            </span>
        </div>

        <div class="toolbar-right">
            <div class="view-switch" role="group" aria-label="顯示模式">
                <button
                    type="button"
                    class:active={viewMode === "code"}
                    aria-pressed={viewMode === "code"}
                    on:click={() => setViewMode("code")}
                >
                    Code
                </button>

                <button
                    type="button"
                    class:active={viewMode === "html"}
                    aria-pressed={viewMode === "html"}
                    on:click={() => setViewMode("html")}
                >
                    HTML
                </button>

                <button
                    type="button"
                    class:active={viewMode === "split"}
                    aria-pressed={viewMode === "split"}
                    on:click={() => setViewMode("split")}
                >
                    Code + HTML
                </button>
            </div>

            <div class="toolbar-actions">
                <button
                    type="button"
                    disabled={isBusy}
                    on:click={() => openFile()}
                >
                    開啟
                </button>

                <button
                    type="button"
                    disabled={isBusy}
                    on:click={() => saveFile()}
                >
                    儲存
                </button>

                <button
                    type="button"
                    disabled={isBusy}
                    on:click={() => saveFile(true)}
                >
                    另存新檔
                </button>
                <button
                  type="button"
                  disabled={isBusy}
                  on:click={() => exportHTML()}
                >
                  匯出 HTML
                </button>
            </div>
        </div>
    </header>

    {#if errorMessage}
        <div class="error-banner" role="alert">
            <span>{errorMessage}</span>

            <button type="button" on:click={() => (errorMessage = "")}>
                關閉
            </button>
        </div>
    {/if}

    {#if externalReadError}
      <div class="external-banner external-error" role="alert">
        <span>
          {externalReadError}。Editor 內容已保留，可使用另存新檔。
        </span>
      </div>
    {/if}

    {#if externalContent !== null}
      <div class="external-banner" role="status">
        {#if confirmingReload}
          <span>
            重新載入會捨棄 Editor 尚未儲存的修改，改用磁碟上的最新內容。
          </span>

          <div class="external-actions">
            <button
              type="button"
              disabled={isBusy}
              on:click={() => reloadFromDisk()}
            >
              確定捨棄並重新載入
            </button>

            <button
              type="button"
              disabled={isBusy}
              on:click={() => (confirmingReload = false)}
            >
              取消
            </button>
          </div>
        {:else}
          <span>
            檔案已被其他程式修改，目前 Editor 也有未儲存的內容。
          </span>

          <div class="external-actions">
            <button
              type="button"
              disabled={isBusy}
              on:click={() => (confirmingReload = true)}
            >
              重新載入
            </button>

            <button
              type="button"
              disabled={isBusy}
              on:click={keepEditorContent}
            >
              保留編輯
            </button>
          </div>
        {/if}
      </div>
    {/if}

    <div
        class="workspace"
        class:mode-code={viewMode === "code"}
        class:mode-html={viewMode === "html"}
        class:mode-split={viewMode === "split"}
    >
        <section class="pane editor-pane" aria-label="Markdown 編輯區">
            <div class="pane-heading">MARKDOWN</div>

            <div class="editor-host" bind:this={editorHost}></div>
        </section>

        <section class="pane preview-pane" aria-label="Markdown 預覽區">
            <div class="pane-heading">PREVIEW</div>

            <div class="preview-scroll">
                <article class="markdown-body">
                    {@html previewHtml}
                </article>
            </div>
        </section>
    </div>

    {#if exportMessage}
      <div class="export-banner" role="status">
        <span>{exportMessage}</span>

        <button
          type="button"
          on:click={() => (exportMessage = '')}
        >
          關閉
        </button>
      </div>
    {/if}

    <footer class="statusbar">
        <span>{lineCount} 行</span>
        <span>{characterCount} 字元</span>
        <span>Ln {cursorLine}, Col {cursorColumn}</span>
        <span class="status-right" title={filePath}>
            {#if isBusy}
                處理中…
            {:else if isDirty}
                尚未儲存
            {:else if filePath}
                已儲存
            {:else}
                尚未儲存至檔案
            {/if}
        </span>
    </footer>
</main>
