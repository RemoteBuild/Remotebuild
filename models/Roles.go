package models

//Role roles for user
type Role struct {
	ID       uint   `gorm:"pk"`
	RoleName string `gorm:"not null"`
	IsAdmin  bool   `gorm:"default:false"`
}

//Permission permission for roles
type Permission uint8

//Permissions
const (
	NoPermission Permission = iota
	ReadPermission
	Writepermission
)
