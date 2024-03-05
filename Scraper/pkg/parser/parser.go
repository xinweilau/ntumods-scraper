package parser

import (
    "encoding/json"
    "github.com/antchfx/htmlquery"
    "golang.org/x/net/html"
    "ntumods/pkg/course"
    "ntumods/pkg/dto"
    "os"
    "strings"
)

func ParseCourses(filePath string) ([]course.Course, error) {
    // Open and parse the HTML file
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    doc, err := html.Parse(file)
    if err != nil {
        return nil, err
    }

    // Query and parse course data
    nodes, err := htmlquery.QueryAll(doc, "//table")
    if err != nil {
        return nil, err
    }

    var courses []course.Course
    for _, node := range nodes {
        c, err := parseToCourse(node)
        if err != nil {
            return nil, err
        }
        courses = append(courses, c)
    }

    return courses, nil
}

func parseToCourse(node *html.Node) (course.Course, error) {
    var resp course.Course

    mappings := map[string]string{
        "Prerequisite:":                         "prerequisite",
        "Grade Type:":                           "grade_type",
        "Mutually exclusive with:":              "mutually_exclusive",
        "Not available to Programme:":           "not_available_to",
        "Not available to all Programme with:":  "not_available_to_prog_with",
        "Not available as BDE/UE to Programme:": "not_available_as_ue",
        "Not available as PE to Programme:":     "not_available_as_pe",
    }

    nodes, err := htmlquery.QueryAll(node, "//font[string-length(text()) > 0]")
    if err != nil {
        return resp, nil
    }

    course := make(map[string]interface{})

    // The first 3 values will always be the course information
    course["code"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[0])), " ")
    course["title"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[1])), " ")
    course["au"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[2])), " ")
    course["description"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[len(nodes)-1])), " ")

    var key string
    for i := 3; i < len(nodes)-1; i += 1 {
        if key == "" {
            key = strings.TrimSpace(htmlquery.InnerText(nodes[i]))
            continue
        }

        value := strings.Join(strings.Fields(htmlquery.InnerText(nodes[i])), " ")

        course[mappings[key]] = value
        key = ""
    }

    // Marshal the interface into JSON
    courseJSON, err := json.Marshal(course)
    if err != nil {
        return resp, err
    }

    // Unmarshal the JSON into a struct
    if err = json.Unmarshal(courseJSON, &resp); err != nil {
        return resp, err
    }
    return resp, nil
}

func ParseCourseSchedules(doc *html.Node) (*dto.CourseSchedules, error) {
    courseSchedules := &dto.CourseSchedules{}

    // Query for the "acadsem" select options
    acadsemNodes, err := htmlquery.QueryAll(doc, `//select[@name="acadsem"]/option[not(contains(@value, "_S"))]`)
    if err != nil {
        return nil, err
    }

    for _, node := range acadsemNodes {
        acadYearSem := htmlquery.SelectAttr(node, "value")
        if acadYearSem != "" {
            courseSchedules.AcadYearSem = append(courseSchedules.AcadYearSem, acadYearSem)
        }
    }

    // Query for the "r_course_yr" select options
    courseYearProgNodes, err := htmlquery.QueryAll(doc, `//select[@name="r_course_yr"]/option`)
    if err != nil {
        return nil, err
    }

    for _, node := range courseYearProgNodes {
        courseYearProg := htmlquery.SelectAttr(node, "value")
        if courseYearProg != "" {
            courseSchedules.CourseYearProg = append(courseSchedules.CourseYearProg, courseYearProg)
        }
    }

    return courseSchedules, nil
}
