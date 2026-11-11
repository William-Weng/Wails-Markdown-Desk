package utility

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// 驗證並開啟使用者拖入的 Markdown 檔案
//
// 參數:
//   - path：拖入檔案的路徑，可能是相對路徑或絕對路徑
//
// 回傳值：
//   - file：已開啟的檔案；呼叫端負責在使用完畢後呼叫 Close()
//   - absolutePath：轉換後的絕對路徑
//   - error：副檔名不符、路徑無效或開檔失敗時回傳錯誤
func OpenMarkdownFile(path string) (*os.File, string, error) {

	extension := strings.ToLower(filepath.Ext(path))

	if extension != ".md" && extension != ".markdown" {
		return nil, "", fmt.Errorf("請拖入 .md 或 .markdown 檔案")
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, "", fmt.Errorf("檔案路徑無效：%w", err)
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return nil, "", fmt.Errorf("開啟檔案失敗：%w", err)
	}

	return file, absolutePath, nil
}

// 讀取指定路徑的文字檔內容
//
// 此函式只接受 UTF-8 編碼的文字檔：
// 1. 讀取整個檔案為位元組資料
// 2. 驗證資料是否為合法 UTF-8
// 3. 若檔案開頭有 UTF-8 BOM，則移除，避免影響 Markdown 第一行的解析
func ReadFileAllContent(path string) (string, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("讀取檔案失敗：%w", err)
	}

	if !utf8.Valid(data) {
		return "", fmt.Errorf("目前僅支援 UTF-8 編碼的文字檔")
	}

	content := strings.TrimPrefix(string(data), "\uFEFF")

	return content, nil
}

// 從已開啟的檔案讀取文字內容
//
// 參數:
//   - file：已開啟的檔案
//   - maxSize：允許讀取的最大內容大小，單位為 byte
//
// 流程：
// 1. 使用 io.LimitReader 限制最多讀取 maxSize + 1 個 byte
// 2. 若實際讀到超過 maxSize，代表檔案太大，拒絕載入
// 3. 驗證內容是否為合法 UTF-8
// 4. 移除 UTF-8 BOM，避免影響 Markdown 第一行解析
func ReadFileContent(file *os.File, maxSize int) (string, error) {

	data, err := ReadFileData(file, maxSize)
	if err != nil {
		return "", err
	}

	content := strings.TrimPrefix(string(data), "\uFEFF")
	return content, nil
}

// 從已開啟的檔案讀取資料，並限制最大讀取大小
//
// 參數:
//   - file：已開啟的檔案；函式不負責關閉它
//   - maxSize：允許讀取的最大資料量，單位為 byte
//
// 讀取策略：
// 使用 io.LimitReader 限制最多讀取 maxSize + 1 個 byte；多讀 1 個 byte 是為了判斷檔案是否「超過」上限：
// - 若檔案剛好等於 maxSize，會讀到 maxSize bytes，允許通過
// - 若檔案大於 maxSize，會讀到 maxSize + 1 bytes，判定為超過上限
//
// 回傳值：
//   - data：讀取到的原始位元組資料
//   - error：讀取失敗，或資料超過上限時回傳錯誤
func ReadFileData(file *os.File, maxSize int) ([]byte, error) {

	data, err := io.ReadAll(io.LimitReader(file, int64(maxSize)+1))
	if err != nil {
		return nil, fmt.Errorf("讀取檔案失敗：%w", err)
	}

	if len(data) > maxSize {
		return nil, fmt.Errorf("目前拖曳開檔限制為 %d bytes", maxSize)
	}

	return data, nil
}

// WriteFile 將文字內容寫入指定路徑
//
// 參數：
//   - path：要寫入的完整檔案路徑
//   - content：要寫入的文字內容
//
// 若檔案不存在，會建立新檔案；若檔案已存在，會先清空原本內容，再寫入新內容；因此這是「覆蓋儲存」，不是附加寫入
func WriteFile(path, content string) error {

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("儲存檔案失敗：%w", err)
	}

	return nil
}

// 取得儲存對話框要預先填入的檔名
//
// 參數：
//   - path：目前文件的路徑；若尚未開啟檔案，通常是空字串
//   - defaultName：沒有現有檔案時要使用的預設檔名
//
// 若已有檔案路徑，沿用其檔名；否則使用 defaultName，例如「Untitled.md」
func GetFilename(path string, defaultaName string) string {

	filename := defaultaName

	if path != "" {
		filename = filepath.Base(path)
	}

	return filename
}

// 檢查已開啟檔案的基本資訊
//
// 參數:
//   - file：已開啟的檔案
//   - maxSize：允許的最大檔案大小，單位為 byte
//
// 檢查項目：
// 1. 能否取得檔案資訊
// 2. 是否為一般檔案，而不是資料夾、裝置或特殊檔案
// 3. 檔案大小是否超過上限
func CheckFileInfo(file *os.File, maxSize int) error {

	info, err := file.Stat()

	if err != nil {
		return fmt.Errorf("取得檔案資訊失敗：%w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("請拖入一般檔案，不是資料夾")
	}

	if info.Size() > int64(maxSize) {
		return fmt.Errorf("目前拖曳開檔限制為 %d bytes", maxSize)
	}

	return nil
}
