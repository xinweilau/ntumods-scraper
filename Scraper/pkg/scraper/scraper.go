package scraper

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"ntumods/pkg/dto"
	"ntumods/pkg/parser"
	"reflect"
	"strings"
	"time"
)

func GetCourseSchedulePair() (*dto.CourseSchedules, error) {
	currYear := time.Now().Year()
	currMonth := time.Now().Month()

	var acadYearSem string
	if currMonth < time.May {
		acadYearSem = fmt.Sprintf("%d_2", currYear-1)
	} else {
		acadYearSem = fmt.Sprintf("%d_1", currYear)
	}

	request := &dto.CourseListRequestDto{
		AcadYearSem: acadYearSem,
	}

	params, err := constructRequiredCourseListFormData(*request)
	if err != nil {
		return nil, err
	}

	resp, err := http.PostForm(dto.CONTENT_OF_COURSES_INIT, *params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return parser.ParseCourseSchedulesList(doc)
}

func GetContentOfCourses(request dto.CourseListRequestDto) ([]dto.Course, error) {
	params, err := constructRequiredCourseListFormData(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.PostForm(dto.CONTENT_OF_COURSES, *params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return parser.ParseCourses(doc)
}

func GetCourseSchedule(request dto.CourseScheduleRequestDto) ([]dto.Module, error) {
	params, err := constructRequiredCourseScheduleFormData(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.PostForm(dto.CLASS_SCHEDULE, *params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return parser.ParseCourseModuleSchedules(doc), nil
}

func constructRequiredCourseListFormData(params dto.CourseListRequestDto) (*url.Values, error) {
	values := &url.Values{}
	vVal := reflect.ValueOf(params)

	requestMap := map[string]string{
		"AcadYearSem": "acadsem",
		"FilterParam": "r_course_yr",
		"SubjectCode": "r_subj_code",
		"BOption":     "boption",
		"AcadYear":    "acad",
		"Semester":    "semester",
	}

	for i := 0; i < vVal.NumField(); i++ {
		field := vVal.Field(i)
		fieldType := vVal.Type().Field(i)
		key := requestMap[fieldType.Name]

		values.Add(key, field.String())
	}

	return values, nil
}

func constructRequiredCourseScheduleFormData(params dto.CourseScheduleRequestDto) (*url.Values, error) {
	values := &url.Values{}
	vVal := reflect.ValueOf(params)

	// The current AcadYearSem uses 2023_1, but the required param here needs 2023;1
	params.AcadYearSem = strings.Replace(params.AcadYearSem, "_", ";", 1)

	requestMap := map[string]string{
		"AcadYearSem": "acadsem",
		"FilterParam": "r_course_yr",
		"SubjectCode": "r_subj_code",
		"BOption":     "boption",
		"SearchType":  "r_search_type",
		"StaffAccess": "staff_access",
	}

	for i := 0; i < vVal.NumField(); i++ {
		field := vVal.Field(i)
		fieldType := vVal.Type().Field(i)
		key := requestMap[fieldType.Name]

		values.Add(key, field.String())
	}

	return values, nil
}
