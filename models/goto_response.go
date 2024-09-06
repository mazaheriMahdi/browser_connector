package models

type GotoResponse struct {
	PageContent string `json:"page_content"`
	PageNumber  int    `json:"page_number"`
}
