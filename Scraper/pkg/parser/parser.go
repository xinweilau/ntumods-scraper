package parser

import (
	"encoding/json"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"ntumods/pkg/dto"
	"strings"
)

func ParseCourses(node *html.Node) ([]dto.Course, error) {
	nodes, err := htmlquery.QueryAll(node, "//table")
	if err != nil {
		return nil, err
	}

	var courses []dto.Course
	for _, node := range nodes {
		c, err := parseToCourse(node)

		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}

	return courses, nil
}

func parseToCourse(node *html.Node) (dto.Course, error) {
	var resp dto.Course

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

func ParseCourseSchedulesList(doc *html.Node) (*dto.CourseSchedules, error) {
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

func ParseCourseModuleSchedules(doc *html.Node) []dto.Module {
	var modules []dto.Module

	tables, _ := htmlquery.QueryAll(doc, "//table")

	for i := 0; i < len(tables); i += 2 {
		if i+1 >= len(tables) {
			break
		}

		firstRowTable := tables[i]
		borderTable := tables[i+1]

		tds, _ := htmlquery.QueryAll(firstRowTable, "./tbody/tr/td")
		if len(tds) < 2 {
			continue
		}
		module := dto.Module{
			Code:  htmlquery.InnerText(tds[0]),
			Title: strings.Trim(htmlquery.InnerText(tds[1]), "*"),
		}

		// Extract schedule details from the border table
		borderRows, _ := htmlquery.QueryAll(borderTable, "./tbody/tr[position()>1]")
		currIndex := ""
		for _, row := range borderRows {
			borderTds, _ := htmlquery.QueryAll(row, "./td")

			var schedule dto.Schedule

			if currIndex != "" {
				time := strings.TrimSpace(htmlquery.InnerText(borderTds[3]))
				rangeOfTime := strings.Split(time, "-")

				schedule = dto.Schedule{
					Index:      currIndex,
					ClassType:  strings.TrimSpace(htmlquery.InnerText(borderTds[0])),
					IndexGroup: strings.TrimSpace(htmlquery.InnerText(borderTds[1])),
					DayOfWeek:  strings.TrimSpace(htmlquery.InnerText(borderTds[2])),
					StartTime:  rangeOfTime[0],
					EndTime:    rangeOfTime[1],
					Venue:      strings.TrimSpace(htmlquery.InnerText(borderTds[4])),
					Remarks:    strings.TrimSpace(htmlquery.InnerText(borderTds[5])),
				}
			} else {
				time := strings.TrimSpace(htmlquery.InnerText(borderTds[4]))
				startTime := ""
				endTime := ""

				if len(time) > 0 {
					rangeOfTime := strings.Split(time, "-")
					startTime = rangeOfTime[0]
					endTime = rangeOfTime[1]
				}

				schedule = dto.Schedule{
					Index:      currIndex,
					ClassType:  strings.TrimSpace(htmlquery.InnerText(borderTds[1])),
					IndexGroup: strings.TrimSpace(htmlquery.InnerText(borderTds[2])),
					DayOfWeek:  strings.TrimSpace(htmlquery.InnerText(borderTds[3])),
					StartTime:  startTime,
					EndTime:    endTime,
					Venue:      strings.TrimSpace(htmlquery.InnerText(borderTds[5])),
					Remarks:    strings.TrimSpace(htmlquery.InnerText(borderTds[6])),
				}
			}

			module.Schedules = append(module.Schedules, schedule)
		}

		currIndex = ""
		modules = append(modules, module)
	}

	return modules
}
