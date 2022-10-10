package models

type GroupModel struct {
	Id      string
	Owner   string
	Image   string
	Name    string
	About   string
	Members []string
}
