package backend

import (
	"MarkdownDesk/backend/configure"
	"MarkdownDesk/backend/model"
)

type ConfigService struct{}

// 回傳前端可使用的應用程式設定
func (service *ConfigService) GetConfig() model.AppConfig {

	return model.AppConfig{
		ImageWebRoute: configure.ImageWebRoute,
	}
}
