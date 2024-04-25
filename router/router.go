package router

import (
	controller "github.com/Aman913k/RecruitmentManagementSystem/controllers"
	"github.com/gorilla/mux"
	
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/signup", controller.Signup).Methods("POST")
	router.HandleFunc("/login", controller.Login).Methods("POST")
	router.HandleFunc("/uploadResume", controller.UploadResume).Methods("POST")
	router.HandleFunc("/admin/job", controller.CreateJob).Methods("POST")

	
	router.HandleFunc("/admin/job/{job_id}", controller.GetJob).Methods("GET")
	router.HandleFunc("/admin/applicants",controller.GetApplicants).Methods("GET")
	router.HandleFunc("/admin/applicant/{applicant_id}", controller.GetApplicant).Methods("GET")
	router.HandleFunc("/jobs", controller.GetJobs).Methods("GET")
	router.HandleFunc("/jobs/apply/{job_id}", controller.ApplyForJob).Methods("POST")

	return router

}
