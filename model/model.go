package model

type User struct {
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	Address         string  `json:"address"`
	UserType        string  `json:"user_type"` // Applicant/Admin
	PasswordHash    string  `json:"password_hash"`
	ProfileHeadline string  `json:"profile_headline"`
	Profile         Profile `json:"profile"`
}

type Profile struct {
	Applicant
	ResumeFileAddress string `json:"resume_file_address"`
}

type Applicant struct {
	Skills     string `json:"skills"`
	Education  string `json:"education"`
	Experience string `json:"experience"`
	Phone      string `json:"phone"`
}

type Job struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	PostedOn          string `json:"posted_on"`
	TotalApplications int    `json:"total_applications"`
	CompanyName       string `json:"company_name"`
	PostedBy          User   `json:"posted_by"`
}

type ResumeDetails struct {
	Education  []Education  `json:"education"`
	Email      string       `json:"email"`
	Experience []Experience `json:"experience"`
	Name       string       `json:"name"`
	Phone      string       `json:"phone"`
	Skills     []string     `json:"skills"`
}

type Education struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Experience struct {
	Dates []string `json:"dates"`
	Name  string   `json:"name"`
	URL   string   `json:"url"`
}
