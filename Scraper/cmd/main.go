package main

import (
    "fmt"
    "ntumods/pkg/parser"
)

func main() {
    courses, err := parser.ParseCourses("../data/index.html")
    if err != nil {
        panic(err)
    }

    fmt.Println(courses[0])
}
