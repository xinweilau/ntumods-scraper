package dto

type Course struct {
	Code                   string `json:"code"`
	Title                  string `json:"title"`
	AU                     string `json:"au"`
	Prerequisite           string `json:"prerequisite"`
	MutuallyExclusive      string `json:"mutually_exclusive"`
	NotAvailableTo         string `json:"not_available_to"`
	NotAvailableToProgWith string `json:"not_available_to_prog_with"`
	GradeType              string `json:"grade_type"`
	NotAvailableAsUE       string `json:"not_available_as_ue"`
	NotAvailableAsPE       string `json:"not_available_as_pe"`
	Description            string `json:"description"`
}
