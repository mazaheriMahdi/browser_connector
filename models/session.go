package models

import (
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
)

type Session struct {
	SessionId uuid.UUID `json:"sessionId"`
	Page      playwright.Page
	isUse     bool
}
