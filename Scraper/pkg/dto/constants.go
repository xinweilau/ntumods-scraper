package dto

import "time"

const (
	CONTENT_OF_COURSES_INIT = "https://wis.ntu.edu.sg/webexe/owa/AUS_SUBJ_CONT.main_display"
	CONTENT_OF_COURSES      = "https://wis.ntu.edu.sg/webexe/owa/AUS_SUBJ_CONT.main_display1"
	CLASS_SCHEDULE          = "https://wis.ntu.edu.sg/webexe/owa/aus_schedule.main_display1"
	EXAM_SCHEDULE           = "https://wis.ntu.edu.sg/pls/webexe/exam_timetable_und.Get_detail"
)

const (
	GET_INITIAL_COURSE_LIST     = "Get Initial Course List"
	GET_COURSE_OFFERED_CONTENTS = "Get Course Offered Contents"
	GET_CLASS_SCHEDULE          = "Get Class Schedule of Course"
	GET_EXAM_SCHEDULE           = "Get Module Exam Schedule"
)

const MAX_RETRIES = 3
const RETRY_DELAY = 5 * time.Second
