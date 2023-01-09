package dualis

import "net/http"

type Dualis struct {
	Client   *http.Client
	Semester []Semester
}

type Semester struct {
	Name    string
	Url     string
	Modules []Module
}

type Module struct {
	Name     string
	Url      string
	Attempts []Attempt
}

type Attempt struct {
	Label  string
	Events []Event
}

type Event struct {
	Name  string
	Grade string
	Exams []Exam
}

type Exam struct {
	Semester string
	Name     string
	Grade    string
}
