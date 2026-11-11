package utility

import "github.com/wailsapp/wails/v3/pkg/application"

// 建立一個「開啟檔案」對話框
//
// 參數：
//   - message：對話框顯示的訊息或標題，例如「開啟 Markdown」
//   - displayName：檔案類型在使用者介面上顯示的名稱，例如「Markdown」
//   - pattern：允許選擇的副檔名模式，例如 "*.md;*.markdown"
//
// 回傳尚未顯示的檔案選擇對話框；呼叫端可再呼叫
func NewMarkdownOpenFileDialog(message string, displayName string, pattern string) *application.OpenFileDialogStruct {

	app := application.Get()
	dialog := app.Dialog.OpenFile().SetMessage(message).AddFilter(displayName, pattern)

	return dialog
}

// 建立一個「儲存 Markdown」的檔案對話框
//
// 參數：
//   - message：對話框顯示的訊息，例如「儲存 Markdown」
//   - filename：對話框預先填入的檔名，例如「Untitled.md」
//   - displayName：檔案類型在使用者介面上的名稱，例如「Markdown」
//   - pattern：允許的副檔名模式，例如 "*.md;*.markdown"
//
// 回傳尚未顯示的儲存檔案對話框；
// 呼叫端必須再呼叫 PromptForSingleSelection() 才會讓使用者選擇。
func NewMarkdownSaveFileDialog(message string, filename string, displayName string, pattern string) *application.SaveFileDialogStruct {
	app := application.Get()
	dialog := app.Dialog.SaveFile().SetMessage(message).SetFilename(filename).AddFilter(displayName, pattern)
	return dialog
}
