package models

import "time"

// add table name to the model
type Folder struct {
	FolderID  uint      `gorm:"primaryKey;autoIncrement" json:"folderId"`
	Name      string    `gorm:"not null" json:"name"`
	OwnerID   uint      `json:"ownerId"`
	Owner     User      `gorm:"foreignKey:OwnerID" json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Folder) TableName() string {
	return "folders"
}

type Note struct {
	NoteID    uint      `gorm:"primaryKey;autoIncrement" json:"noteId"`
	Title     string    `gorm:"not null" json:"title"`
	Body      string    `json:"body"`
	FolderID  uint      `json:"folderId"`
	Folder    Folder    `gorm:"foreignKey:FolderID" json:"folder"`
	OwnerID   uint      `json:"ownerId"`
	Owner     User      `gorm:"foreignKey:OwnerID" json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type FolderShare struct {
	FolderID uint   `gorm:"primaryKey" json:"folderId"`
	UserID   uint   `gorm:"primaryKey" json:"userId"`
	Access   string `gorm:"not null" json:"access"` // "read" or "write"
}

type NoteShare struct {
	NoteID uint   `gorm:"primaryKey" json:"noteId"`
	UserID uint   `gorm:"primaryKey" json:"userId"`
	Access string `gorm:"not null" json:"access"` // "read" or "write"
}
