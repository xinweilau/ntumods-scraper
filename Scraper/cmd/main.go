package main

import (
    "fmt"
    "ntumods/pkg/parser"
    "ntumods/pkg/scraper"
)

func main() {
    _, err := parser.ParseCourses("../data/index.html")
    if err != nil {
        panic(err)
    }

    init, err := scraper.GetCourseSchedulePair()
    if err != nil {
        panic(err)
    }

    fmt.Println(init.AcadYearSem[len(init.AcadYearSem)-1])
}
