package model

import "time"

// Repo is a starred repository (fields needed for classify + render).
type Repo struct {
	FullName        string
	HTMLURL         string
	Description     string
	Language        string
	Stars           int
	Topics          []string
	Archived        bool
	StarredAt       time.Time
	Homepage        string
}

// Category is a dynamic group under one user.
type Category struct {
	Name  string
	Repos []Repo
}

// UserStars is one upstream user's classified stars.
type UserStars struct {
	Username   string
	Label      string
	Categories []Category
	Total      int
	Err        string // non-empty if fetch failed for this user
}

// Report is the full render input.
type Report struct {
	Title     string
	Generated time.Time
	Users     []UserStars
	Total     int
}
