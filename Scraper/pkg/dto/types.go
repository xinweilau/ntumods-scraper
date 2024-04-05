package dto

type Course struct {
	Code                   string  `json:"code"`
	Title                  string  `json:"title"`
	AU                     string  `json:"au"`
	Prerequisite           string  `json:"prerequisite"`
	MutuallyExclusive      string  `json:"mutually_exclusive"`
	NotAvailableTo         string  `json:"not_available_to"`
	NotAvailableToProgWith string  `json:"not_available_to_prog_with"`
	GradeType              string  `json:"grade_type"`
	NotAvailableAsUE       string  `json:"not_available_as_ue"`
	NotAvailableAsPE       string  `json:"not_available_as_pe"`
	Description            string  `json:"description"`
	Faculty                Faculty `json:"faculty"`
	NotOfferedAsBDE        bool    `json:"notOfferedAsBDE"`
}

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

type CourseExamScheduleRequestDto struct {
	ExamSubject     string
	PlanNo          string
	ExamDateTime    string
	ExamStartTime   string
	ExamDepartment  string
	ExamVenue       string
	Matric          string
	AcademicSession string
	ExamYear        string
	ExamSemester    string
	ExamType        string
	BOption         string
}

type CourseSchedules struct {
	AcadYearSem    []string
	CourseYearProg []string
}

// ModuleLite is a lightweight representation of a module
type ModuleLite struct {
	Code        string  `json:"code"`
	Module      string  `json:"module"`
	Description string  `json:"description"`
	AU          string  `json:"au"`
	Faculty     Faculty `json:"faculty"`
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

type ExamSchedule struct {
	Date      string `json:"date"`
	DayOfWeek string `json:"dayOfWeek"`
	Time      string `json:"time"`
	Code      string `json:"code"`
	Title     string `json:"title"`
	Duration  string `json:"duration"`
}

type Faculty struct {
	Title string `json:"Faculty"`
	Code  string `json:"Code"`
}
