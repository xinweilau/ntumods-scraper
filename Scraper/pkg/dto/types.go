package dto

type CourseListRequestDto struct {
    AcadYearSem string
    FilterParam string
    SubjectCode string
    BOption     string
    AcadYear    string
    Semester    string
}

type CourseSchedules struct {
    AcadYearSem    []string
    CourseYearProg []string
}
