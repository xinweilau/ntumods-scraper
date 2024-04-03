package dto

type CourseListRequestDto struct {
	AcadYearSem string
	FilterParam string
	SubjectCode string
	BOption     string
	AcadYear    string
	Semester    string
}

type CourseScheduleRequestDto struct {
	AcadYearSem string
	FilterParam string
	SubjectCode string
	BOption     string
	SearchType  string
	StaffAccess string
}

type CourseSchedules struct {
	AcadYearSem    []string
	CourseYearProg []string
}

// ModuleLite is a lightweight representation of a module
type ModuleLite struct {
	Code   string `json:"code"`
	Module string `json:"module"`
}

// Module is a structure containing the module code, title, and the semesters which it is offered
type Module struct {
	Code      string
	Title     string
	Schedules []Schedule `json:"schedules"`
}

type Schedule struct {
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Venue      string `json:"venue"`
	ClassType  string `json:"classType"`
	Index      string `json:"index"`
	IndexGroup string `json:"indexGroup"`
	DayOfWeek  string `json:"dayOfWeek"`
	Remarks    string `json:"remarks"`
}
