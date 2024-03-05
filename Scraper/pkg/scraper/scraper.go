package scraper

import (
    "fmt"
    "golang.org/x/net/html"
    "net/http"
    "net/url"
    "ntumods/pkg/dto"
    "ntumods/pkg/parser"
    "reflect"
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

    params, err := constructRequiredFormData(*request)
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

    return parser.ParseCourseSchedules(doc)
}

func GetContentOfCourses(url string, request dto.CourseListRequestDto) (*html.Node, error) {
    params, err := constructRequiredFormData(request)
    if err != nil {
        return nil, err
    }

    resp, _ := http.PostForm(url, *params)
    defer resp.Body.Close()

    doc, err := html.Parse(resp.Body)
    return doc, nil
}

func constructRequiredFormData(params dto.CourseListRequestDto) (*url.Values, error) {
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
