package models

type ProfileParams struct {
	Name           string `json:"name" binding:"required"`
	NativeLanguage string `json:"native_language" binding:"required"`
	TargetLanguage string `json:"target_language" binding:"required"`
}
