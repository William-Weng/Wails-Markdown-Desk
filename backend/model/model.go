package model

// 代表一份已開啟或已儲存的 Markdown 文件
type DocumentFile struct {
	Path    string `json:"path"`    // 文件的完整磁碟路徑
	Name    string `json:"name"`    // 檔名
	Content string `json:"content"` // 文件的文字內容
}

// 是要提供給 Web 前端的應用程式設定
type AppConfig struct {
	ImageWebRoute string `json:"imageWebRoute"` // 圖片起始路由
}
