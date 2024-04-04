package dto

import "reflect"

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
	Code        string `json:"code"`
	Module      string `json:"module"`
	Description string `json:"description"`
	AU          string `json:"au"`
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

func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !IsEmpty(v.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}
