package entity

import (
	"encoding/json"
	"time"
)

type UserGroup struct {
	ID          int       `gorm:"autoIncrement" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	NamePy      string    `json:"namePy"`
	Description string    `json:"description"`
	Tags        string    `json:"tags"`
}

func (c *UserGroup) MarshalJSON() ([]byte, error) {
	type Alias UserGroup
	return json.Marshal(&struct {
		*Alias
		CreatedAt DateTime `json:"createdAt"`
	}{
		(*Alias)(c),
		DateTime(c.CreatedAt),
	})
}
