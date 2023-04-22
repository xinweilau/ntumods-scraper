package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)


type Course struct {
	Code string `json:"code"`
	Title string `json:"title"`
	AU string `json:"au"`
	Prerequisite string `json:"prerequisite"`
	MutuallyExclusive string `json:"mutually_exclusive"`
	NotAvailableTo string `json:"not_available_to"`
	NotAvailableToProgWith string `json:"not_available_to_prog_with"`
	GradeType string `json:"grade_type"`
	NotAvailableAsUE string `json:"not_available_as_ue"`
	NotAvailableAsPE string `json:"not_available_as_pe"`
	Description string `json:"description"`
}

func getHTMLExample() (*html.Node, error) {
	file, err := os.Open("index.html")
    if err != nil {
        return nil, err
    }
    defer file.Close()

    doc, err := html.Parse(file)
    if err != nil {
        return nil, err
    }

    return doc, nil
}

func main() {
	doc, err := getHTMLExample()
	if err != nil {
		panic(err)
	}

	nodes, err := htmlquery.QueryAll(doc, "//table")
	if err != nil {
		panic(err)
	}

	courses := make([]Course, 0)
	for _, node := range nodes {
		course, err := parseToCourse(node)
		if err != nil {
			panic(err)
		}

		courses = append(courses, course)
	}

	fmt.Println(courses[0].NotAvailableTo)
}

func parseToCourse(node *html.Node) (Course, error) {
	var resp Course

	mappings := map[string]string{
		"Prerequisite:": "prerequisite",
		"Grade Type:": "grade_type",
		"Mutually exclusive with:": "mutually_exclusive",
		"Not available to Programme:": "not_available_to",
		"Not available to all Programme with:": "not_available_to_prog_with",
		"Not available as BDE/UE to Programme:": "not_available_as_ue",
		"Not available as PE to Programme:": "not_available_as_pe",
	}

	nodes, err := htmlquery.QueryAll(node, "//font[string-length(text()) > 0]")
	if err != nil {
		return resp, nil
	}

	course := make(map[string]interface{})

	// The first 3 values will always be the course information
	course["code"] = htmlquery.InnerText(nodes[0])
	course["title"] = htmlquery.InnerText(nodes[1])
	course["au"] = htmlquery.InnerText(nodes[2])
	course["description"] = htmlquery.InnerText(nodes[len(nodes)-1])

	var key string
	for i := 3; i < len(nodes) - 1; i += 1 {
		if key == "" {
			key = strings.TrimSpace(htmlquery.InnerText(nodes[i]))
			continue
		}

		value := strings.TrimSpace(htmlquery.InnerText(nodes[i]))

		course[mappings[key]] = value
		key = ""
	}

	// Marshal the interface into JSON
	courseJSON, err := json.Marshal(course)
	if err != nil {
		return resp, err
	}

	// Unmarhsal the JSON into a struct
	if err = json.Unmarshal(courseJSON, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}