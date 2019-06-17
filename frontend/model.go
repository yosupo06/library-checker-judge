package main

import (
	"database/sql"
	"html/template"
)

type Problem struct {
	Name      string
	Title     string
	Statement template.HTML
}

type Submission struct {
	ID        int
	Problem   string
	Lang      string
	Status    string
	Source    string
	MaxTime   int
	MaxMemory int
	UserID    sql.NullInt64
	User      User
}

type Task struct {
	Submission int
}

type SubmissionTestcaseResult struct {
	Submission int
	Testcase   string
	Status     string
	Time       int
	Memory     int
}

type User struct {
	ID       int
	Name     string
	Passhash string
}

func (u User) getID() sql.NullInt64 {
	if u.Name == "" {
		return sql.NullInt64{0, false}
	}
	return sql.NullInt64{int64(u.ID), true}
}
