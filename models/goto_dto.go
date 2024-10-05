package models

type GotoDto struct {
	Url        string `json:"url"`
	PageHeight int    `json:"pageHeight"`
	PageWidth  int    `json:"pageWidth"`
}
