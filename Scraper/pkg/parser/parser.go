package parser

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"ntumods/pkg/dto"
	"ntumods/pkg/utils"
	"strconv"
	"strings"
)

func ParseCourses(node *html.Node) ([]dto.Course, error) {
	nodes, err := htmlquery.QueryAll(node, "//table")
	if err != nil {
		return nil, err
	}

	numChild, err := htmlquery.QueryAll(node, "//table/tbody/tr[count(*) > 3]")
	if err != nil {
		return nil, err
	}

	var courses []dto.Course

	// Minor Courses and BDEs
	if len(numChild) > 0 {
		courses, err = parseMinorAndBDE(nodes[0])
		if err != nil {
			return nil, err
		}
	} else {
		// Normal Courses
		for _, n := range nodes {
			c, err := parseNormalCourse(n)

			if err != nil {
				return nil, err
			}
			courses = append(courses, c)
		}
	}

	return courses, nil
}

func parseMinorAndBDE(node *html.Node) ([]dto.Course, error) {
	var resp []dto.Course

	mappings := map[string]string{
		"Prerequisite:":                         "prerequisite",
		"Grade Type:":                           "grade_type",
		"Mutually exclusive with:":              "mutually_exclusive",
		"Not available to Programme:":           "not_available_to",
		"Not available to all Programme with:":  "not_available_to_prog_with",
		"Not available as BDE/UE to Programme:": "not_available_as_ue",
		"Not available as PE to Programme:":     "not_available_as_pe",
	}

	nodes, err := htmlquery.QueryAll(node, "//table//tr[position() > 1]/td")
	if err != nil {
		return resp, nil
	}

	if len(nodes) == 0 {
		return resp, nil
	}

	courses := make([]map[string]interface{}, 0)

	course := make(map[string]interface{})
	key := ""
	isMultiRow := false
	isMetaData := true

	// need to modify to skip the 4th item
	for i := 0; i < len(nodes)-1; i += 1 {
		innerText := htmlquery.InnerText(nodes[i])
		value := strings.Join(strings.Fields(innerText), " ")

		// Blank space for other values
		if innerText == "" {
			continue
		}

		// blank space between rows
		if len(innerText) > 0 && strings.TrimSpace(innerText) == "" {
			courses = append(courses, course)
			course = make(map[string]interface{})
			isMetaData = true
			continue
		}

		if isMetaData {
			course["code"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[i])), " ")
			course["title"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[i+1])), " ")

			// 3.0 AU -> 3
			rawAU := strings.Join(strings.Fields(htmlquery.InnerText(nodes[i+2])), " ")
			rawAU = strings.Split(rawAU, " ")[0]
			res, err := strconv.ParseFloat(rawAU, 32)
			if err != nil {
				fmt.Println("[parseMinorAndBDE] Error parsing AU for", course["code"])
			}
			course["au"] = res

			// dept is a non-existent key, exist only for completeness and does not get translated to course struct later on
			course["dept"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[i+3])), " ")

			i += 3
			isMetaData = false
			continue
		}

		if _, exists := mappings[value]; exists {
			key = value
			continue
		}

		if key != "" {
			isMultiRow = strings.HasSuffix(value, "OR")

			if curr, exist := course[mappings[key]]; exist {
				course[mappings[key]] = curr.(string) + " " + value
			} else {
				course[mappings[key]] = value
			}

			if !isMultiRow {
				key = ""
			}
			continue
		}

		course["description"] = value

		// if this is the last course, then it needs to be added
		if i == len(nodes)-2 {
			courses = append(courses, course)
			course = make(map[string]interface{})
			isMetaData = true
		}
	}

	for _, c := range courses {
		// Marshal the interface into JSON
		courseJSON, err := json.Marshal(c)
		if err != nil {
			return resp, err
		}

		var respCourse dto.Course
		// Unmarshal the JSON into a struct
		if err = json.Unmarshal(courseJSON, &respCourse); err != nil {
			return resp, err
		}

		resp = append(resp, respCourse)
	}

	return resp, nil
}

func parseNormalCourse(node *html.Node) (dto.Course, error) {
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

	if len(nodes) == 0 {
		return resp, nil
	}

	course := make(map[string]interface{})

	// The first 3 values will always be the course information
	course["code"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[0])), " ")
	course["title"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[1])), " ")

	// 3.0 AU -> 3
	rawAU := strings.Join(strings.Fields(htmlquery.InnerText(nodes[2])), " ")
	rawAU = strings.Split(rawAU, " ")[0]
	res, err := strconv.ParseFloat(rawAU, 32)
	if err != nil {
		fmt.Println("[parseNormalCourse] Error parsing AU for", course["code"])
	}
	course["au"] = res

	course["description"] = strings.Join(strings.Fields(htmlquery.InnerText(nodes[len(nodes)-1])), " ")

	opt := strings.Join(strings.Fields(htmlquery.InnerText(nodes[len(nodes)-2])), " ")
	course["notOfferedAsBDE"] = opt == "Not offered as Broadening and Deepening Elective"

	var key string
	isMultiRow := false

	for i := 3; i < len(nodes)-1; i += 1 {
		if key == "" {
			key = strings.TrimSpace(htmlquery.InnerText(nodes[i]))
			continue
		}

		value := strings.Join(strings.Fields(htmlquery.InnerText(nodes[i])), " ")

		isMultiRow = strings.HasSuffix(value, "OR")

		if curr, exist := course[mappings[key]]; exist {
			course[mappings[key]] = curr.(string) + " " + value
		} else {
			course[mappings[key]] = value
		}

		if !isMultiRow {
			key = ""
		}
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
		index := ""
		for _, row := range borderRows {
			borderTds, _ := htmlquery.QueryAll(row, "./td")
			currIndex := strings.TrimSpace(htmlquery.InnerText(borderTds[0]))

			if currIndex != "" {
				index = currIndex
			}

			var schedule dto.Schedule
			time := strings.TrimSpace(htmlquery.InnerText(borderTds[4]))
			startTime := ""
			endTime := ""

			if len(time) > 0 {
				rangeOfTime := strings.Split(time, "-")
				startTime = rangeOfTime[0]
				endTime = rangeOfTime[1]
			}

			teachingWeeks := make([]int, 0)
			remarks := strings.TrimSpace(htmlquery.InnerText(borderTds[6]))
			// Teaching Wk1-4,6-9,11-13
			if strings.Contains(remarks, "Teaching Wk") {
				weeks := strings.Split(remarks, "Teaching Wk")[1]
				rangeWeeks := strings.Split(weeks, ",")

				for _, r := range rangeWeeks {
					startEnd := strings.Split(r, "-")
					start, _ := strconv.Atoi(startEnd[0])

					// means that it is not a range, but a singular week
					if len(startEnd) == 1 {
						teachingWeeks = append(teachingWeeks, start)
						continue
					}

					end, _ := strconv.Atoi(startEnd[1])

					sliceRange := utils.CreateIntSlice(start, end)
					teachingWeeks = append(teachingWeeks, sliceRange...)
				}

				if len(teachingWeeks) == 0 {
					teachingWeeks = utils.CreateIntSlice(1, 13)
				}
			} else if remarks != "Not conducted during Teaching Weeks" {
				sliceRange := utils.CreateIntSlice(1, 13)
				teachingWeeks = append(teachingWeeks, sliceRange...)
			}

			schedule = dto.Schedule{
				Index:         index,
				ClassType:     strings.TrimSpace(htmlquery.InnerText(borderTds[1])),
				IndexGroup:    strings.TrimSpace(htmlquery.InnerText(borderTds[2])),
				DayOfWeek:     strings.TrimSpace(htmlquery.InnerText(borderTds[3])),
				StartTime:     startTime,
				EndTime:       endTime,
				Venue:         strings.TrimSpace(htmlquery.InnerText(borderTds[5])),
				Remarks:       strings.TrimSpace(htmlquery.InnerText(borderTds[6])),
				TeachingWeeks: teachingWeeks,
			}

			module.Schedules = append(module.Schedules, schedule)
		}

		modules = append(modules, module)
	}

	return modules
}
func ParseExamSchedules(doc *html.Node) ([]dto.ExamSchedule, error) {
	examSchedule := make([]dto.ExamSchedule, 0)

	examNodes, err := htmlquery.QueryAll(doc, `//table[@border="1"]/tbody/tr[not(td/@colspan="7") and normalize-space(td)]`)
	if err != nil {
		return nil, err
	}

	// Iterate through each row of exam schedule
	// Skips first row as it is the header of the table
	if len(examNodes) > 1 {
		examNodes = examNodes[1:]
	}

	for _, node := range examNodes {
		scheduleNode := htmlquery.Find(node, "./td")

		examSchedule = append(examSchedule, dto.ExamSchedule{
			Date:      strings.Join(strings.Fields(htmlquery.InnerText(scheduleNode[0])), " "),
			DayOfWeek: strings.Join(strings.Fields(htmlquery.InnerText(scheduleNode[1])), " "),
			Time:      strings.Join(strings.Fields(htmlquery.InnerText(scheduleNode[2])), " "),
			Code:      strings.Join(strings.Fields(htmlquery.InnerText(scheduleNode[3])), " "),
			Title:     strings.Join(strings.Fields(htmlquery.InnerText(scheduleNode[4])), " "),
			Duration:  strings.Join(strings.Fields(htmlquery.InnerText(scheduleNode[5])), " "),
		})
	}

	return examSchedule, nil
}
