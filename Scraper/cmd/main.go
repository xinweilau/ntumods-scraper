package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"ntumods/pkg/dto"
	"ntumods/pkg/scraper"
	"ntumods/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxWorkers = 3

var processedCourses sync.Map
var moduleList []dto.ModuleLite

type courseDetailParams struct {
	AcadYearSem    string
	CourseYearProg string
}

type examDetailParams struct {
	AcadYearSem string
	Code        string
}

type Combined struct {
	dto.Course
	Exam     dto.ExamSchedule `json:"exam"`
	Schedule []dto.Schedule   `json:"schedule"`
}

var courseDetailWg sync.WaitGroup
var examDetailWg sync.WaitGroup

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	facultyInformation, err := populateFacultyInformation()
	if err != nil {
		panic(err)
	}

	init, err := scraper.GetCourseSchedulePair()
	if err != nil {
		panic(err)
	}

	courseYearProgChan := make(chan courseDetailParams, maxWorkers)
	courseChan := make(chan courseDetailParams, maxWorkers)
	examChan := make(chan examDetailParams, maxWorkers*2) // Double in size as it is a bottleneck; there will be a lot of courses extracted from getCourseTimetable.

	// Start worker A goroutines
	for i := 0; i < maxWorkers; i++ {
		courseDetailWg.Add(1)
		go getContentOfCourses(courseYearProgChan, facultyInformation)
	}

	// Start worker B goroutines
	for i := 0; i < maxWorkers; i++ {
		courseDetailWg.Add(1)
		go getCourseTimetable(courseChan, examChan)
	}

	// Start worker C goroutines
	for i := 0; i < maxWorkers*2; i++ {
		examDetailWg.Add(1)
		go getExamSchedule(examChan)
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

		request := courseDetailParams{
			CourseYearProg: courseYearProg,
			AcadYearSem:    latestSemester,
		}

		courseYearProgChan <- request
		courseChan <- request
	}

	close(courseYearProgChan)
	close(courseChan)
	courseDetailWg.Wait()

	close(examChan)
	examDetailWg.Wait()

	numModules := 0
	processedCourses.Range(func(key, value interface{}) bool {
		// For some reason there's always an empty Course.json generated, this is to bypass that
		if key == "Course" {
			return true
		}

		c := value.(Combined)

		moduleLite := dto.ModuleLite{
			Code:        c.Course.Code,
			Module:      c.Course.Title,
			AU:          c.AU,
			Description: c.Description,
			Faculty:     c.Course.Faculty,
		}

		if !utils.IsEmpty(moduleLite) {
			moduleList = append(moduleList, moduleLite)
		}

		numModules += 1

		fileName := key.(string)
		blobName := filepath.Join(latestSemester, fileName+".json")
		if err = utils.UploadFileToBlobStorage(blobName, value); err != nil {
			fmt.Println("Error uploading file to blob storage:", err)
			return false
		}
		return true
	})

	blobName := filepath.Join(latestSemester, "moduleList.json")
	if err = utils.UploadFileToBlobStorage(blobName, moduleList); err != nil {
		fmt.Println("Error uploading file to blob storage:", err)
	}

	fmt.Println("Extraction Complete (numModules = ", numModules, ")")
}

func getContentOfCourses(courseYearProgChan <-chan courseDetailParams, facultyInformation map[string]dto.Faculty) {
	defer courseDetailWg.Done() // Decrement the counter when the goroutine completes
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
			fmt.Println("Error in getContentOfCourses:", err)
			continue
		}

		for _, c := range res {
			if utils.IsEmpty(c) {
				continue
			}

			// Need to resolve the course code here, course code could start with either 2 or 3 characters
			// We will first attempt to map 2 characters, because it could be possible AA might supersede AAA
			var faculty dto.Faculty
			codeA := c.Code[:2]
			if f, exists := facultyInformation[codeA]; exists {
				faculty = f
			}

			codeB := c.Code[:3]
			if f, exists := facultyInformation[codeB]; exists {
				faculty = f
			}

			c.Faculty = faculty

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

func getCourseTimetable(courseChan <-chan courseDetailParams, examChan chan<- examDetailParams) {
	defer courseDetailWg.Done() // Decrement the counter when the goroutine completes
	for course := range courseChan {
		fmt.Println("[WorkerB] Processing Course Schedule (", course.AcadYearSem, ", ", course.CourseYearProg, ")")
		request := dto.CourseScheduleRequestDto{
			AcadYearSem: course.AcadYearSem,
			FilterParam: course.CourseYearProg,
			BOption:     "CLoad",
		}

		res, err := scraper.GetCourseSchedule(request)
		if err != nil {
			fmt.Println("Error in getCourseTimetable:", err)
			continue
		}

		for _, c := range res {
			if utils.IsEmpty(c) {
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
			examChan <- examDetailParams{
				AcadYearSem: course.AcadYearSem,
				Code:        c.Code,
			}
		}
	}
}

func getExamSchedule(examChan <-chan examDetailParams) {
	defer examDetailWg.Done() // Decrement the counter when the goroutine completes
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
			fmt.Println("Error in getExamSchedule:", err)
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

func populateFacultyInformation() (map[string]dto.Faculty, error) {
	data, err := os.ReadFile("../data/faculty.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	var facultyData map[string]dto.Faculty
	err = json.Unmarshal(data, &facultyData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil, err
	}

	output := make(map[string]dto.Faculty)
	for keys, faculty := range facultyData {
		for _, key := range strings.Split(keys, ";") {
			output[key] = faculty
		}
	}

	return output, nil
}
