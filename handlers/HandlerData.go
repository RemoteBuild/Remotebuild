package handlers

import (
	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/jinzhu/gorm"
)

//HandlerData handlerData for web
type HandlerData struct {
	Config *models.Config
	Db     *gorm.DB
	User   *models.User
}
