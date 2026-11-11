package utility

import (
	"MarkdownDesk/backend/configure"
	"MarkdownDesk/backend/model"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/net/html"
)

// 檢查 HTTP 請求方法是否只允許 GET 或 HEAD
//
// GET：取得圖片內容
// HEAD：取得回應標頭，但不回傳圖片內容
//
// 其他方法，例如 POST、PUT、DELETE，回傳 405 Method Not Allowed
//
// 參數:
//   - request: 目前的 HTTP 請求
//
// 回傳值:
//   - 方法允許時回傳 nil
//   - 方法不允許時回傳 *HttpError，狀態碼為 405
func ValidRequestMethodStandard(request *http.Request) error {
	switch request.Method {
	case http.MethodGet, http.MethodHead:
		return nil
	default:
		return &model.HttpError{Message: "Method not allowed", Code: http.StatusMethodNotAllowed}
	}
}

// 解析 Markdown 中引用的圖片路徑，並確保它只能是「相對於目前文件」的路徑
//
// 允許的例子：
//   - "images/logo.png"
//   - "./assets/chart.png"
//   - "sub/dir/photo.jpg"
//
// 拒絕的例子：
//   - "https://example.com/a.png"：外部 URL
//   - "/etc/passwd"：絕對路徑
//   - "C:\\Windows\\file.png"：Windows 絕對路徑
//   - ""：空路徑
//
// 參數:
//   - source: Markdown 圖片語法中的 src 值
//
// 回傳值:
//   - 成功時，回傳可交給 os.Root.Open 的相對路徑
//   - 失敗時，回傳帶有 400 或 403 的 *HttpError
func ParseImagePath(source string) (string, error) {

	_url, err := url.Parse(source)
	if err != nil {
		return "", &model.HttpError{Message: "Invalid image path", Code: http.StatusBadRequest}
	}

	if _url.Scheme != "" || _url.Host != "" || _url.Opaque != "" {
		return "", &model.HttpError{Message: "Relative paths only", Code: http.StatusForbidden}
	}

	imagePath := filepath.FromSlash(_url.Path)

	if imagePath == "" ||
		filepath.IsAbs(imagePath) ||
		filepath.VolumeName(imagePath) != "" {
		return "", &model.HttpError{Message: "Invalid image path", Code: http.StatusForbidden}
	}

	return imagePath, nil
}

// WriteHttpError 將錯誤寫入 HTTP 回應
//
// 若錯誤是 *HttpError，使用其自訂的 HTTP 狀態碼；若是一般 error，則以 500 Internal Server Error 回應
//
// 參數:
//   - writer: HTTP 回應寫入器
//   - err: 要回傳給客戶端的錯誤；nil 時不寫入任何回應
func WriteHttpError(writer http.ResponseWriter, err error) {

	if err == nil {
		return
	}

	if httpErr, ok := err.(*model.HttpError); ok {
		http.Error(writer, httpErr.Error(), httpErr.Code)
	} else {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// NewImageFile 在目前 Markdown 文件所在資料夾內，安全地開啟圖片檔案
//
// 參數：
//   - currentPath：目前 Markdown 文件的完整路徑
//   - imagePath：相對於文件所在資料夾的圖片路徑
//
// 安全機制：
// 使用 os.OpenRoot 把檔案存取限制在文件所在資料夾內，避免透過 ..、絕對路徑或符號連結讀取資料夾以外的檔案
//
// 回傳值：
//   - file：已開啟的圖片檔案；呼叫端負責關閉
//   - info：圖片檔案的 metadata，例如名稱、大小、修改時間
//   - error：開啟失敗、路徑不允許或不是一般檔案時回傳錯誤
func NewImageFile(currentPath string, imagePath string) (*os.File, os.FileInfo, error) {

	base, err := filepath.Abs(filepath.Dir(currentPath))

	if err != nil {
		return nil, nil, &model.HttpError{Message: "Invalid document directory", Code: http.StatusInternalServerError}
	}

	root, err := os.OpenRoot(base)
	if err != nil {
		return nil, nil, &model.HttpError{Message: "Cannot open document directory", Code: http.StatusNotFound}
	}
	defer root.Close()

	file, err := root.Open(imagePath)
	if err != nil {
		return nil, nil, &model.HttpError{Message: "Image not found or not allowed", Code: http.StatusNotFound}
	}

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, nil, &model.HttpError{Message: "Invalid image file", Code: http.StatusNotFound}
	}

	return file, info, nil
}

// ParseContentType 讀取圖片開頭最多 512 bytes，並依內容特徵判斷 MIME type
//
// http.DetectContentType 會根據檔案 magic bytes 判斷格式，例如 PNG、JPEG、GIF 都有明確的檔頭特徵
//
// 參數:
//   - file: 已開啟的圖片檔案
//
// 回傳值:
//   - 成功時，回傳 MIME type，例如 "image/png"
//   - 檔案為空或無法讀取時，回傳 415 錯誤
func ParseContentType(file *os.File) (string, error) {

	var header [512]byte
	number, _ := file.Read(header[:])

	if number == 0 {
		return "", &model.HttpError{Message: "Empty or unreadable image", Code: http.StatusUnsupportedMediaType}
	}

	contentType := http.DetectContentType(header[:number])
	return contentType, nil
}

// 將文件內的圖片複製到指定的資源目錄；若目標圖片檔已存在，則直接重用舊檔案，不會重新複製
//
// 參數：
//   - documentPath：來源文件路徑
//   - source      ：文件中記錄的圖片來源
//   - assetDir    ：匯出圖片要存放的目錄
//   - index       ：圖片編號，用來產生穩定的輸出檔名
//
// 回傳值：
//   - string：成功時回傳輸出檔名，例如 "image-3.png"
//   - error ：解析路徑、判斷格式或複製圖片失敗時回傳錯誤
func CopyExportImage(documentPath string, source string, assetDir string, index int) (string, error) {

	imagePath, err := ParseImagePath(source)
	if err != nil {
		return "", err
	}

	input, _, err := NewImageFile(documentPath, imagePath)
	if err != nil {
		return "", err
	}

	contentType, err := ParseContentType(input)
	if err != nil {
		return "", err
	}

	extension, err := parseExtension(contentType)
	if err != nil {
		return "", err
	}

	if _, err := input.Seek(0, 0); err != nil {
		return "", err
	}

	name := fmt.Sprintf("image-%d%s", index, extension)

	output, err := os.OpenFile(filepath.Join(assetDir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}

	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()

	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}

	return name, nil
}

// 建立一個 HTML 節點巡訪函式，用來將 <img> 的內部圖片 URL 還原成 Markdown 原本記錄的相對路徑
//
// 匯出 HTML 時，文件內的圖片可能使用應用程式內部 URL，例如：
//
//	/document-images?document=/path/to/note.md&src=images/photo.png
//
// 這個函式會將它還原成：
//
//	images/photo.png
//
// 如此一來，匯出的 HTML 不需要複製圖片，
// 只要和原始 Markdown、圖片維持相同的相對位置，瀏覽器就能正常顯示圖片。
//
// 參數：
//   - documentPath：目前開啟的來源文件完整路徑，例如 /Users/name/Documents/note.md
//
// 回傳值：
//   - func(*html.Node) error： 可傳入 html.Parse 後的根節點，递迴巡訪整份 HTML 並修改 <img src>
func NewImagePathRestorer(documentPath string) func(*html.Node) error {

	var walk func(*html.Node) error

	walk = func(node *html.Node) error {

		if node.Type == html.ElementNode && node.Data == "img" {

			for index := range node.Attr {
				attr := &node.Attr[index]
				if attr.Key != "src" {
					continue
				}

				imageURL, err := url.Parse(attr.Val)
				if err != nil {
					return fmt.Errorf("圖片網址無效：%w", err)
				}

				if imageURL.Scheme == "https" || imageURL.Scheme == "http" {
					continue
				}

				if imageURL.Scheme != "" || imageURL.Host != "" || imageURL.Path != configure.ImageWebRoute {
					return fmt.Errorf("不支援的圖片網址：%s", attr.Val)
				}

				query := imageURL.Query()

				if documentPath == "" || query.Get("document") != documentPath {
					return fmt.Errorf("本機圖片的來源文件不符")
				}

				attr.Val = query.Get("src")
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child); err != nil {
				return err
			}
		}

		return nil
	}

	return walk
}

// 依據圖片的 MIME type 回傳對應副檔名
func parseExtension(mime string) (string, error) {

	switch mime {
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	case "image/gif":
		return ".gif", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", fmt.Errorf("不支援的圖片格式")
	}
}
