package main

import "html/template"

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
	Maxtime   int
	Maxmemory int
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
