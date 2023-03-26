package entity

type Config struct {
	Id       int    `gorm:"autoIncrement" json:"id"`
	ItemName string `json:"itemName"`
	Content  string `json:"content"`
}
