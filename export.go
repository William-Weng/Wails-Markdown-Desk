package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/net/html"
)

// ExportHTML 匯出完整 HTML，並複製其中的本機圖片。
// 取消對話框時回傳空字串。
func (s *FileService) ExportHTML(
	documentHTML string,
	documentPath string,
) (string, error) {
	s.mu.Lock()
	currentPath := s.currentPath
	s.mu.Unlock()

	if documentPath != currentPath {
		return "", fmt.Errorf("目前文件已切換，請重新匯出")
	}

	filename := "Untitled.html"

	if documentPath != "" {
		name := filepath.Base(documentPath)
		filename = strings.TrimSuffix(name, filepath.Ext(name)) + ".html"
	}

	outputPath, err := application.Get().Dialog.SaveFile().
		SetMessage("匯出 HTML").
		SetFilename(filename).
		AddFilter("HTML", "*.html;*.htm").
		PromptForSingleSelection()

	if err != nil {
		return "", err
	}
	if outputPath == "" {
		return "", nil
	}

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

		if sourceErr == nil && outputErr == nil &&
			os.SameFile(sourceInfo, outputInfo) {
			return "", fmt.Errorf("不能覆蓋目前的 Markdown 來源文件")
		}
	}

	doc, err := html.Parse(strings.NewReader(documentHTML))
	if err != nil {
		return "", fmt.Errorf("解析匯出 HTML 失敗：%w", err)
	}

	outputDir := filepath.Dir(outputPath)
	baseName := strings.TrimSuffix(
		filepath.Base(outputPath),
		filepath.Ext(outputPath),
	)

	var assetDir string
	success := false

	// 失敗時，只清理這次新建的圖片資料夾。
	defer func() {
		if !success && assetDir != "" {
			_ = os.RemoveAll(assetDir)
		}
	}()

	copied := make(map[string]string)
	imageCount := 0

	var walk func(*html.Node) error

	walk = func(node *html.Node) error {
		if node.Type == html.ElementNode && node.Data == "img" {
			for index := range node.Attr {
				attr := &node.Attr[index]

				if attr.Key != "src" {
					continue
				}

				u, err := url.Parse(attr.Val)
				if err != nil {
					return fmt.Errorf("圖片網址無效：%w", err)
				}

				// 網路圖片保持原樣，不在匯出時下載。
				if u.Scheme == "https" || u.Scheme == "http" {
					continue
				}

				if u.Scheme != "" || u.Host != "" ||
					u.Path != "/document-images" {
					return fmt.Errorf("不支援的圖片網址：%s", attr.Val)
				}

				query := u.Query()

				if documentPath == "" ||
					query.Get("document") != documentPath {
					return fmt.Errorf("本機圖片的來源文件不符")
				}

				source := query.Get("src")

				if cached, ok := copied[source]; ok {
					attr.Val = cached
					continue
				}

				if assetDir == "" {
					assetDir, err = os.MkdirTemp(
						outputDir,
						baseName+"-assets-",
					)
					if err != nil {
						return fmt.Errorf("建立圖片資料夾失敗：%w", err)
					}
				}

				imageCount++

				imageName, err := copyExportImage(
					documentPath,
					source,
					assetDir,
					imageCount,
				)
				if err != nil {
					return fmt.Errorf("匯出圖片 %q 失敗：%w", source, err)
				}

				// 轉成 URL path，處理資料夾名稱中的空白或中文。
				relativePath := filepath.ToSlash(
					filepath.Join(filepath.Base(assetDir), imageName),
				)

				imageURL := (&url.URL{
					Path: relativePath,
				}).String()

				copied[source] = imageURL
				attr.Val = imageURL
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(doc); err != nil {
		return "", err
	}

	var output strings.Builder

	if err := html.Render(&output, doc); err != nil {
		return "", fmt.Errorf("產生 HTML 失敗：%w", err)
	}

	if err := os.WriteFile(
		outputPath,
		[]byte(output.String()),
		0o644,
	); err != nil {
		return "", fmt.Errorf("寫入 HTML 失敗：%w", err)
	}

	success = true
	return outputPath, nil
}

// copyExportImage 以串流複製圖片，不轉成 Base64。
func copyExportImage(
	documentPath string,
	source string,
	assetDir string,
	index int,
) (string, error) {
	u, err := url.Parse(source)
	if err != nil {
		return "", err
	}

	if u.Scheme != "" || u.Host != "" || u.Opaque != "" {
		return "", fmt.Errorf("僅允許相對圖片路徑")
	}

	imagePath := filepath.FromSlash(u.Path)

	if imagePath == "" ||
		filepath.IsAbs(imagePath) ||
		filepath.VolumeName(imagePath) != "" {
		return "", fmt.Errorf("圖片路徑無效")
	}

	base, err := filepath.Abs(filepath.Dir(documentPath))
	if err != nil {
		return "", err
	}

	root, err := os.OpenRoot(base)
	if err != nil {
		return "", err
	}
	defer root.Close()

	input, err := root.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer input.Close()

	info, err := input.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("圖片不是一般檔案")
	}

	var header [512]byte
	n, _ := input.Read(header[:])

	var extension string

	switch http.DetectContentType(header[:n]) {
	case "image/png":
		extension = ".png"
	case "image/jpeg":
		extension = ".jpg"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	default:
		return "", fmt.Errorf("不支援的圖片格式")
	}

	if _, err := input.Seek(0, 0); err != nil {
		return "", err
	}

	name := fmt.Sprintf("image-%d%s", index, extension)

	output, err := os.OpenFile(
		filepath.Join(assetDir, name),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o644,
	)
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
