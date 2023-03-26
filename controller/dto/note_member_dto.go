package dto

type NoteShareDto struct {
	Id        int    `json:"id"`
	NoteId    int    `json:"noteId"`
	ShareType string `json:"shareType"`
	Role      int    `json:"role"`
}

type NoteUnShareDto struct {
	Id        int    `json:"id"`
	NoteId    int    `json:"noteId"`
	ShareType string `json:"shareType"`
}

type GroupRoleDto struct {
	UserId int `json:"userId"`
	Role   int `json:"role"`
}
