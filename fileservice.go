package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type DocumentFile struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type FileService struct {
	mu          sync.Mutex
	currentPath string
}

// Open 顯示對話框並讀取文件。
// 使用者取消時回傳 nil, nil，不視為錯誤。
func (s *FileService) Open() (*DocumentFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := application.Get().Dialog.OpenFile().
		SetMessage("開啟 Markdown").
		AddFilter("Markdown", "*.md;*.markdown").
		PromptForSingleSelection()

	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取檔案失敗：%w", err)
	}

	if !utf8.Valid(data) {
		return nil, fmt.Errorf("目前僅支援 UTF-8 編碼的文字檔")
	}

	// 移除 UTF-8 BOM，避免影響第一行 Markdown 的解析。
	content := strings.TrimPrefix(string(data), "\uFEFF")

	s.currentPath = path

	return &DocumentFile{
		Path:    path,
		Name:    filepath.Base(path),
		Content: content,
	}, nil
}

// Save 儲存到目前文件；尚無路徑或 saveAs 為 true 時顯示另存對話框。
func (s *FileService) Save(
	content string,
	saveAs bool,
) (*DocumentFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.currentPath

	if saveAs || path == "" {
		filename := "Untitled.md"
		if path != "" {
			filename = filepath.Base(path)
		}

		selected, err := application.Get().Dialog.SaveFile().
			SetMessage("儲存 Markdown").
			SetFilename(filename).
			AddFilter("Markdown", "*.md;*.markdown").
			PromptForSingleSelection()

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

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("儲存檔案失敗：%w", err)
	}

	// 只有寫入成功，才更新目前的文件路徑。
	s.currentPath = path

	return &DocumentFile{
		Path: path,
		Name: filepath.Base(path),
	}, nil
}

// ReadImage 只允許讀取目前 Markdown 所在資料夾內的圖片。
// documentPath 用來確認前端預覽與後端目前開啟的是同一份文件。
func (s *FileService) ReadImage(
	documentPath string,
	source string,
) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentPath == "" {
		return "", fmt.Errorf("請先儲存或開啟 Markdown 檔案")
	}

	if documentPath != s.currentPath {
		return "", fmt.Errorf("文件已切換，請重新產生預覽")
	}

	u, err := url.Parse(source)
	if err != nil {
		return "", fmt.Errorf("圖片路徑無效")
	}

	// 網路圖片由前端載入；這裡不接受 file: 或其他 scheme。
	if u.Scheme != "" || u.Host != "" || u.Opaque != "" {
		return "", fmt.Errorf("僅允許文件內的相對圖片路徑")
	}

	imagePath := filepath.FromSlash(u.Path)

	if imagePath == "" || filepath.IsAbs(imagePath) ||
		filepath.VolumeName(imagePath) != "" {
		return "", fmt.Errorf("請使用相對圖片路徑")
	}

	base, err := filepath.Abs(filepath.Dir(s.currentPath))
	if err != nil {
		return "", err
	}

	// Go 1.24+：把檔案存取限制在文件資料夾內。
	// os.Root 也會限制透過符號連結逃離目錄。
	root, err := os.OpenRoot(base)
	if err != nil {
		return "", fmt.Errorf("開啟文件資料夾失敗：%w", err)
	}
	defer root.Close()

	file, err := root.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("無法開啟圖片：%w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("圖片必須是一般檔案")
	}

	const maxImageSize = 10 * 1024 * 1024

	if info.Size() > maxImageSize {
		return "", fmt.Errorf("圖片超過 10 MiB")
	}

	data, err := io.ReadAll(io.LimitReader(file, maxImageSize+1))
	if err != nil {
		return "", fmt.Errorf("讀取圖片失敗：%w", err)
	}
	if len(data) > maxImageSize {
		return "", fmt.Errorf("圖片超過 10 MiB")
	}

	mimeType := http.DetectContentType(data)

	switch mimeType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
	default:
		return "", fmt.Errorf("目前只支援 PNG、JPEG、GIF、WebP")
	}

	return "data:" + mimeType + ";base64," +
		base64.StdEncoding.EncodeToString(data), nil
}

// ReadCurrent 重新讀取目前開啟的文件。
// 不顯示開檔對話框，也不切換目前文件。
func (s *FileService) ReadCurrent(
	expectedPath string,
) (*DocumentFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentPath == "" {
		return nil, nil
	}

	// 前端等待回覆期間可能切換文件，因此先確認路徑。
	if s.currentPath != expectedPath {
		return nil, nil
	}

	data, err := os.ReadFile(s.currentPath)
	if err != nil {
		return nil, fmt.Errorf("重新讀取檔案失敗：%w", err)
	}

	if !utf8.Valid(data) {
		return nil, fmt.Errorf("目前僅支援 UTF-8 編碼的文字檔")
	}

	return &DocumentFile{
		Path:    s.currentPath,
		Name:    filepath.Base(s.currentPath),
		Content: strings.TrimPrefix(string(data), "\uFEFF"),
	}, nil
}

// ServeHTTP 提供目前 Markdown 文件內的圖片。
// 此服務會掛在 /document-images 路由。
func (s *FileService) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	documentPath := r.URL.Query().Get("document")
	source := r.URL.Query().Get("src")

	// 只短暫鎖定以取得目前文件路徑，
	// 不要在整個圖片傳輸期間持有鎖。
	s.mu.Lock()
	currentPath := s.currentPath
	s.mu.Unlock()

	if currentPath == "" || documentPath != currentPath {
		http.Error(w, "Document is not active", http.StatusForbidden)
		return
	}

	u, err := url.Parse(source)
	if err != nil {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	if u.Scheme != "" || u.Host != "" || u.Opaque != "" {
		http.Error(w, "Relative paths only", http.StatusForbidden)
		return
	}

	imagePath := filepath.FromSlash(u.Path)

	if imagePath == "" ||
		filepath.IsAbs(imagePath) ||
		filepath.VolumeName(imagePath) != "" {
		http.Error(w, "Invalid image path", http.StatusForbidden)
		return
	}

	base, err := filepath.Abs(filepath.Dir(currentPath))
	if err != nil {
		http.Error(w, "Invalid document directory", http.StatusInternalServerError)
		return
	}

	// 限制圖片只能位於目前文件資料夾內。
	root, err := os.OpenRoot(base)
	if err != nil {
		http.Error(w, "Cannot open document directory", http.StatusNotFound)
		return
	}
	defer root.Close()

	file, err := root.Open(imagePath)
	if err != nil {
		http.Error(w, "Image not found or not allowed", http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "Invalid image file", http.StatusNotFound)
		return
	}

	// 只讀取前 512 bytes 判斷格式，不把整張圖片讀進記憶體。
	var header [512]byte
	n, _ := file.Read(header[:])

	if n == 0 {
		http.Error(w, "Empty or unreadable image", http.StatusUnsupportedMediaType)
		return
	}

	contentType := http.DetectContentType(header[:n])

	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
	default:
		http.Error(w, "Unsupported image format", http.StatusUnsupportedMediaType)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "Cannot read image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// 允許快取，但重用之前要驗證。
	w.Header().Set("Cache-Control", "private, no-cache")

	http.ServeContent(
		w,
		r,
		info.Name(),
		info.ModTime(),
		file,
	)
}
