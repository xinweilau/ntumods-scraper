package main

import (
    "encoding/json"
    "fmt"
    "ntumods/pkg/dto"
    "ntumods/pkg/scraper"
    "os"
    "path/filepath"
    "strings"
    "sync"
)

const maxWorkers = 3

var processedCourses sync.Map
var moduleList []dto.ModuleLite

type waParams struct {
    AcadYearSem    string
    CourseYearProg string
}

type wbParams struct {
    AcadYearSem string
    Code        string
}

type Combined struct {
    dto.Course
    Exam     dto.ExamSchedule `json:"exam"`
    Schedule []dto.Schedule   `json:"schedule"`
}

var wg sync.WaitGroup
var wg2 sync.WaitGroup

func main() {
    init, err := scraper.GetCourseSchedulePair()
    if err != nil {
        panic(err)
    }

    courseYearProgChan := make(chan waParams, maxWorkers)
    courseChan := make(chan waParams, maxWorkers)
    examChan := make(chan wbParams, maxWorkers)

    // Start worker A goroutines
    for i := 0; i < maxWorkers; i++ {
        wg.Add(1) // Increment the WaitGroup counter before starting the goroutine
        go workerA(courseYearProgChan)
    }

    // Start worker B goroutines
    for i := 0; i < maxWorkers; i++ {
        wg.Add(1) // Do the same for workerB goroutines
        go workerB(courseChan, examChan)
    }

    // Start worker C goroutines
    for i := 0; i < maxWorkers; i++ {
        wg2.Add(1) // Do the same for workerB goroutines
        go workerC(examChan)
    }

    numSemesters := len(init.AcadYearSem)
    latestSemester := init.AcadYearSem[numSemesters-1]
    for i := numSemesters - 1; i >= 0; i-- {
        semester := init.AcadYearSem[i]
        if !strings.Contains(semester, "Special Term") {
            latestSemester = semester
            break
        }
    }

    // Send CourseYearProg data to worker A and worker B goroutines
    numCourses := len(init.CourseYearProg)
    for i := 0; i < numCourses; i++ {
        courseYearProg := init.CourseYearProg[i]

        request := waParams{
            CourseYearProg: courseYearProg,
            AcadYearSem:    latestSemester,
        }

        courseYearProgChan <- request
        //courseChan <- request
    }

    close(courseYearProgChan)
    close(courseChan)
    wg.Wait()

    close(examChan) // Close after worker B has finished
    wg2.Wait()      // Wait for worker C to finish

    numModules := 0
    processedCourses.Range(func(key, value interface{}) bool {
        c := value.(Combined)

        moduleLite := dto.ModuleLite{
            Code:   c.Course.Code,
            Module: c.Course.Title,
        }
        moduleList = append(moduleList, moduleLite)

        numModules += 1

        exportStructToFile(key.(string), value)
        return true
    })

    exportStructToFile("moduleList", moduleList)

    fmt.Println("Extraction Complete (numModules = ", numModules, ")")
}

func workerA(courseYearProgChan <-chan waParams) {
    defer wg.Done() // Decrement the counter when the goroutine completes
    for courseYearProg := range courseYearProgChan {
        fmt.Println("[WorkerA] Processing Course Content (", courseYearProg.AcadYearSem, ", ", courseYearProg.CourseYearProg, ")")

        acadYear := strings.Split(courseYearProg.AcadYearSem, "_")[0]
        semester := strings.Split(courseYearProg.AcadYearSem, "_")[1]

        request := dto.CourseListRequestDto{
            AcadYearSem: courseYearProg.AcadYearSem,
            FilterParam: courseYearProg.CourseYearProg,
            BOption:     "CLoad",
            AcadYear:    acadYear,
            Semester:    semester,
        }

        res, err := scraper.GetContentOfCourses(request)
        if err != nil {
            fmt.Println("Error in workerA:", err)
            continue
        }

        for _, c := range res {
            if dto.IsEmpty(c) {
                continue
            }

            if loaded, exists := processedCourses.Load(c.Code); exists {
                if currCombined, ok := loaded.(Combined); ok {
                    processedCourses.Store(c.Code, Combined{
                        Course:   c,
                        Exam:     currCombined.Exam,
                        Schedule: currCombined.Schedule,
                    })
                }
            } else {
                processedCourses.Store(c.Code, Combined{
                    Course: c,
                })
            }
        }
    }
}

func workerB(courseChan <-chan waParams, examChan chan<- wbParams) {
    defer wg.Done() // Decrement the counter when the goroutine completes
    for course := range courseChan {
        fmt.Println("[WorkerB] Processing Course Schedule (", course.AcadYearSem, ", ", course.CourseYearProg, ")")
        request := dto.CourseScheduleRequestDto{
            AcadYearSem: course.AcadYearSem,
            FilterParam: course.CourseYearProg,
            BOption:     "CLoad",
        }

        res, err := scraper.GetCourseSchedule(request)
        if err != nil {
            fmt.Println("Error in workerB:", err)
            continue
        }

        for _, c := range res {
            if dto.IsEmpty(c) {
                continue
            }

            if loaded, exists := processedCourses.Load(c.Code); exists {
                if currCombined, ok := loaded.(Combined); ok {
                    processedCourses.Store(c.Code, Combined{
                        Course:   currCombined.Course,
                        Exam:     currCombined.Exam,
                        Schedule: c.Schedules,
                    })
                }
            } else {
                processedCourses.Store(c.Code, Combined{
                    Schedule: c.Schedules,
                })
            }
        }

        for _, c := range res {
            if dto.IsEmpty(c) {
                continue
            }

            examChan <- wbParams{
                AcadYearSem: course.AcadYearSem,
                Code:        c.Code,
            }
        }
    }
}

func workerC(examChan <-chan wbParams) {
    defer wg2.Done() // Decrement the counter when the goroutine completes
    for course := range examChan {
        fmt.Println("[WorkerC] Processing Exam Schedule (", course.AcadYearSem, ", ", course.Code, ")")
        acadSem := strings.Split(course.AcadYearSem, "_")

        request := dto.CourseExamScheduleRequestDto{
            ExamSemester: acadSem[1],
            ExamYear:     acadSem[0],
            BOption:      "Next",
            PlanNo:       "110",
            ExamType:     "UE",
            ExamSubject:  course.Code,
        }

        res, err := scraper.GetExamSchedule(request)
        if err != nil {
            fmt.Println("Error in workerC:", err)
            continue
        }

        for _, exam := range res {
            if loaded, exists := processedCourses.Load(exam.Code); exists {
                if currCombined, ok := loaded.(Combined); ok {
                    processedCourses.Store(exam.Code, Combined{
                        Course:   currCombined.Course,
                        Schedule: currCombined.Schedule,
                        Exam:     exam,
                    })
                }
            } else {
                processedCourses.Store(exam.Code, Combined{
                    Exam: exam,
                })
            }
        }
    }
}

func exportStructToFile(filename string, data interface{}) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error marshaling data:", err)
        return
    }

    filePath := filepath.Join("..", "out", filename+".json")

    if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
        err := os.Mkdir(filepath.Dir(filePath), os.ModePerm)
        if err != nil {
            fmt.Println("Error creating directory:", err)
            return
        }
    }

    file, err := os.Create(filePath)
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }
    defer file.Close()

    _, err = file.Write(jsonData)
    if err != nil {
        fmt.Println("Error writing to file:", err)
        return
    }
}
