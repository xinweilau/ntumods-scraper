package scraper

import (
	"fmt"
	"golang.org/x/net/html"
	"math"
	"math/rand"
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

	resp, err := postFormWithExponentialBackoff(dto.GET_INITIAL_COURSE_LIST, dto.CONTENT_OF_COURSES_INIT, *params)
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

	resp, err := postFormWithExponentialBackoff(dto.GET_COURSE_OFFERED_CONTENTS, dto.CONTENT_OF_COURSES, *params)
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

	resp, err := postFormWithExponentialBackoff(dto.GET_CLASS_SCHEDULE, dto.CLASS_SCHEDULE, *params)
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

func GetExamSchedule(request dto.CourseExamScheduleRequestDto) ([]dto.ExamSchedule, error) {
	params, err := constructRequiredExamScheduleFormData(request)
	if err != nil {
		return nil, err
	}

	res, err := httpPostWithExponentialBackoff(dto.GET_EXAM_SCHEDULE, dto.EXAM_SCHEDULE, *params)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	if err != nil {
		return nil, err
	}

	return parser.ParseExamSchedules(doc)
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

func constructRequiredExamScheduleFormData(params dto.CourseExamScheduleRequestDto) (*url.Values, error) {
	values := &url.Values{}
	vVal := reflect.ValueOf(params)

	requestMap := map[string]string{
		"ExamSubject":     "p_subj",
		"PlanNo":          "p_plan_no",
		"ExamDateTime":    "p_exam_dt",
		"ExamStartTime":   "p_start_time",
		"ExamDepartment":  "p_dept",
		"ExamVenue":       "p_venue",
		"Matric":          "p_matric",
		"AcademicSession": "academic_session",
		"ExamYear":        "p_exam_yr",
		"ExamSemester":    "p_semester",
		"ExamType":        "p_type",
		"BOption":         "bOption",
	}

	for i := 0; i < vVal.NumField(); i++ {
		field := vVal.Field(i)
		fieldType := vVal.Type().Field(i)
		key := requestMap[fieldType.Name]

		values.Add(key, field.String())
	}

	return values, nil
}

func postFormWithExponentialBackoff(service string, url string, data url.Values) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 0; attempt < dto.MAX_RETRIES; attempt++ {
		resp, err = http.PostForm(url, data)
		if err == nil {
			return resp, nil
		}

		// Calculate the delay with exponential backoff and some randomness
		delay := dto.RETRY_DELAY*time.Duration(math.Pow(2, float64(attempt))) + time.Duration(rand.Intn(int(dto.RETRY_DELAY)))
		fmt.Printf("failed to fetch data for [%d], will retry in [%d] seconds\n", service, delay)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("after %d attempts, last error: %s", dto.MAX_RETRIES, err)
}

func httpPostWithExponentialBackoff(service string, url string, data url.Values) (*http.Response, error) {
	client := &http.Client{}

	var resp *http.Response
	var err error
	for attempt := 0; attempt < dto.MAX_RETRIES; attempt++ {
		payload := strings.NewReader(data.Encode())
		req, err := http.NewRequest("POST", url, payload)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err = client.Do(req)
		if err == nil {
			return resp, nil
		}

		// Calculate the delay with exponential backoff and some randomness
		delay := dto.RETRY_DELAY*time.Duration(math.Pow(2, float64(attempt))) + time.Duration(rand.Intn(int(dto.RETRY_DELAY)))
		fmt.Printf("failed to fetch data for [%d], will retry in [%d] seconds\n", service, delay)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("after %d attempts, last error: %s", dto.MAX_RETRIES, err)
}
