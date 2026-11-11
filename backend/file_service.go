package backend

import (
	"MarkdownDesk/backend/configure"
	"MarkdownDesk/backend/model"
	"MarkdownDesk/backend/utility"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// FileService 負責管理 Markdown 文件的開啟、儲存、重新讀取、圖片存取與 HTML 匯出
type FileService struct {
	mutex       sync.Mutex // 避免多個 goroutine 同時讀寫 currentPath
	currentPath string     // 目前開啟的 Markdown 文件完整路徑
}

// Open 讓使用者選擇一個 Markdown 檔案，讀取其內容後回傳代表該文件的 DocumentFile
//
// 若使用者取消選擇，回傳 (nil, nil)；
// 若選檔、讀檔或編碼驗證失敗，回傳 (nil, error)
func (service *FileService) Open() (*model.DocumentFile, error) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	dialog := utility.NewMarkdownOpenFileDialog("開啟 Markdown", "Markdown", configure.MarkdownFilter)
	path, err := dialog.PromptForSingleSelection()

	if err != nil {
		return nil, err
	}

	if path == "" {
		return nil, nil
	}

	content, err := utility.ReadFileAllContent(path)
	if err != nil {
		return nil, err
	}

	service.currentPath = path
	return &model.DocumentFile{Path: path, Name: filepath.Base(path), Content: content}, nil
}

// 儲存目前的 Markdown 文件
//
// 參數:
//   - content：要儲存的文件內容
//   - saveAs：是否強制顯示「另存新檔」對話框
//
// 行為：
// - saveAs 為 true，或尚未開啟任何檔案時，顯示儲存對話框
// - 使用者取消儲存時，回傳 (nil, nil)
// - 寫入成功後，更新 service.currentPath
// - 寫入失敗時，不更新 currentPath，並回傳錯誤
func (service *FileService) Save(content string, saveAs bool) (*model.DocumentFile, error) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	path := service.currentPath

	if saveAs || path == "" {

		filename := utility.GetFilename(path, fmt.Sprintf("%s.md", configure.DefaultFileName))
		dialog := utility.NewMarkdownSaveFileDialog("儲存 Markdown", filename, "Markdown", configure.MarkdownFilter)
		selected, err := dialog.PromptForSingleSelection()

		if err != nil {
			return nil, err
		}
		if selected == "" {
			return nil, nil
		}

		path = selected

		if filepath.Ext(path) == "" {
			path += ".md"
		}
	}

	if err := utility.WriteFile(path, content); err != nil {
		return nil, fmt.Errorf("儲存檔案失敗：%w", err)
	}

	service.currentPath = path
	return &model.DocumentFile{Path: path, Name: filepath.Base(path)}, nil
}

// 重新讀取目前開啟的 Markdown 文件
//
// 只有當 service.currentPath 與 expectedPath 相同時，才會重新讀取；這可避免以下情況：呼叫端仍持有舊文件的路徑，但 FileService 已切換到另一份文件；此時若直接讀取，可能會把錯誤文件的內容覆蓋到畫面上
//
// 回傳值：
// - 目前沒有開啟文件：回傳 (nil, nil)
// - expectedPath 與目前文件路徑不一致：回傳 (nil, nil)
// - 讀取或 UTF-8 驗證失敗：回傳 (nil, error)
// - 成功：回傳包含最新內容的 DocumentFile
//
// 參數:
//   - expectedPath: 呼叫端預期的文件完整路徑
func (service *FileService) ReadCurrent(expectedPath string) (*model.DocumentFile, error) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.currentPath == "" {
		return nil, nil
	}

	if service.currentPath != expectedPath {
		return nil, nil
	}

	content, err := utility.ReadFileAllContent(service.currentPath)
	if err != nil {
		return nil, err
	}

	return &model.DocumentFile{Path: service.currentPath, Name: filepath.Base(service.currentPath), Content: content}, nil
}

// 依使用者拖入的路徑開啟 Markdown 文件
//
// 參數:
//   - path：拖入檔案的路徑
//
// 流程：
// 1. 驗證副檔名並開啟檔案
// 2. 檢查是否為一般檔案，以及大小是否超過上限
// 3. 讀取內容、驗證 UTF-8，並移除 BOM
// 4. 更新目前文件路徑
// 5. 回傳 DocumentFile
func (service *FileService) OpenPath(path string) (*model.DocumentFile, error) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	file, absolutePath, err := utility.OpenMarkdownFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = utility.CheckFileInfo(file, configure.MaxDocumentSize)
	if err != nil {
		return nil, err
	}

	content, err := utility.ReadFileContent(file, configure.MaxDocumentSize)
	if err != nil {
		return nil, err
	}

	service.currentPath = absolutePath
	return &model.DocumentFile{Path: absolutePath, Name: filepath.Base(absolutePath), Content: content}, nil
}

// 提供 /document-images 路由的 HTTP handler
//
// 查詢參數：
//   - document：目前 Markdown 文件的完整路徑
//   - src：Markdown 內引用的相對圖片路徑
//
// 範例：
// GET /document-images?document=/notes/readme.md&src=images/logo.png
//
// 安全設計：
// 1. 只允許 GET 與 HEAD
// 2. 必須指定目前開啟的 Markdown 文件
// 3. 圖片必須是相對路徑
// 4. 使用 os.OpenRoot 將存取限制在文件資料夾內
// 5. 只允許 PNG、JPEG、GIF、WebP
// 6. 回應不允許瀏覽器自行猜測 Content-Type
func (service *FileService) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	err := utility.ValidRequestMethodStandard(request)
	if err != nil {
		utility.WriteHttpError(writer, err)
		return
	}

	documentPath := request.URL.Query().Get("document")
	source := request.URL.Query().Get("src")

	service.mutex.Lock()
	currentPath := service.currentPath
	service.mutex.Unlock()

	if currentPath == "" || documentPath != currentPath {
		http.Error(writer, "Document is not active", http.StatusForbidden)
		return
	}

	imagePath, err := utility.ParseImagePath(source)
	if err != nil {
		utility.WriteHttpError(writer, err)
		return
	}

	file, info, err := utility.NewImageFile(currentPath, imagePath)
	if err != nil {
		utility.WriteHttpError(writer, err)
		return
	}
	defer file.Close()

	contentType, err := utility.ParseContentType(file)
	if err != nil {
		utility.WriteHttpError(writer, err)
		return
	}

	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
	default:
		http.Error(writer, "Unsupported image format", http.StatusUnsupportedMediaType)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		http.Error(writer, "Cannot read image", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Cache-Control", "private, no-cache")

	http.ServeContent(writer, request, info.Name(), info.ModTime(), file)
}

// 將目前的 HTML 內容匯出成檔案
//
// 本機圖片不會被複製或轉檔，HTML 會保留文件中原本記錄的圖片來源；因此輸出 HTML 必須和原始 Markdown 及圖片維持相同的相對位置
func (service *FileService) ExportHTML(documentHTML string, documentPath string) (string, error) {

	service.mutex.Lock()
	currentPath := service.currentPath
	service.mutex.Unlock()

	if documentPath != currentPath {
		return "", fmt.Errorf("目前文件已切換，請重新匯出")
	}

	filename := fmt.Sprintf("%s.html", configure.DefaultFileName)

	if documentPath != "" {
		name := filepath.Base(documentPath)
		filename = strings.TrimSuffix(name, filepath.Ext(name)) + ".html"
	}

	dialog := utility.NewMarkdownSaveFileDialog("匯出 HTML", filename, "HTML", configure.HtmlFilter)
	outputPath, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if outputPath == "" {
		return "", nil
	}

	// 檢查輸出副檔名。
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == "" {
		outputPath += ".html"
	} else if ext != ".html" && ext != ".htm" {
		return "", fmt.Errorf("請使用 .html 或 .htm 副檔名")
	}

	// 不允許把已開啟的來源文件直接當成輸出目的地。
	if documentPath != "" {
		sourceInfo, sourceErr := os.Stat(documentPath)
		outputInfo, outputErr := os.Stat(outputPath)

		if sourceErr == nil && outputErr == nil && os.SameFile(sourceInfo, outputInfo) {
			return "", fmt.Errorf("不能覆蓋目前的 Markdown 來源文件")
		}
	}

	// 解析目前編輯器產生的 HTML。
	doc, err := html.Parse(strings.NewReader(documentHTML))
	if err != nil {
		return "", fmt.Errorf("解析匯出 HTML 失敗：%w", err)
	}

	var restoreImagePaths = utility.NewImagePathRestorer(documentPath)
	if err := restoreImagePaths(doc); err != nil {
		return "", err
	}

	var output strings.Builder

	if err := html.Render(&output, doc); err != nil {
		return "", fmt.Errorf("產生 HTML 失敗：%w", err)
	}

	// 寫入使用者選擇的輸出檔案。
	if err := utility.WriteFile(outputPath, output.String()); err != nil {
		return "", err
	}

	return outputPath, nil
}
