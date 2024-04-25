package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Aman913k/RecruitmentManagementSystem/model"
	"github.com/gorilla/mux"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JWTClaims struct {
	Email    string `json:"email"`
	UserType string `json:"userType"`
	jwt.StandardClaims
}

var JWTSecret = []byte("your-secret-key")


const connectionString = "mongodb+srv://amanrana9133:Aman1n1@cluster0.6ohqxgo.mongodb.net/students_db?retryWrites=true&w=majority"
const dbName = "RecruitmentMS"
const colName = "RMS"

var collection *mongo.Collection

// Initialize MongoDB client and database
func init() {

	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Mongo Connection Successful")
	collection = client.Database(dbName).Collection(colName)
}


func Signup(w http.ResponseWriter, r *http.Request) {
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Name == "" || user.Email == "" || user.PasswordHash == "" || user.UserType == "" || user.ProfileHeadline == "" || user.Address == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Insert user into MongoDB
	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User profile created successfully"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	var foundUser model.User
	err = collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		fmt.Println("hi")
		return
	}

	// Validate password
	if user.PasswordHash != foundUser.PasswordHash {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generating JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		Email: foundUser.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		},
	})
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// Handles uploading resume to the 3rd-party API for extraction
func UploadResume(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) 
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Reading file content
	resumeContent, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Creating  a new HTTP client
	client := &http.Client{}

	apiEndpoint := "https://api.apilayer.com/resume_parser/upload"

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(resumeContent))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}


	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("apikey", "gNiXyflsFu3WNYCz1ZCxdWDb7oQg1Nl1")

	
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, string(respBody), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

// Handles creating job openings
func CreateJob(w http.ResponseWriter, r *http.Request) {
	var job model.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := job.PostedBy.Email
	if email == "" {
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	// Retrieve user from the database based on email
	var user model.User
	err = collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		http.Error(w, "Unauthorized: User not found", http.StatusUnauthorized)
		return
	}

	// Check if user is an admin
	if user.UserType != "Admin" {
		http.Error(w, "Forbidden: Only admins can access this resource", http.StatusForbidden)
		return
	}

	
	job.PostedOn = time.Now().Format("2006-01-02") // Set the posted on date to current date
	job.PostedBy.Name = user.Name                 
	_, err = collection.InsertOne(context.Background(), job)
	if err != nil {
		http.Error(w, "Failed to insert job into database", http.StatusInternalServerError)
		return
	}

	
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Job created successfully: %+v", job)
}

// Fetching information regarding a job opening
func GetJob(w http.ResponseWriter, r *http.Request) {
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var jobs []model.Job
	for cursor.Next(context.Background()) {
		var job model.Job
		err := cursor.Decode(&job)
		if err != nil {
			http.Error(w, "Failed to decode job", http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
	}

	// Return list of jobs in response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// Fetching a list of all users in the system
func GetApplicants(w http.ResponseWriter, r *http.Request) {
	userType := r.Header.Get("userType")
	if userType == "" {
		http.Error(w, "User type not provided in headers", http.StatusBadRequest)
		return
	}

	// Check if the user type is Admin
	if userType != "Admin" {
		http.Error(w, "Forbidden: Only admins can access this resource", http.StatusForbidden)
		return
	}

	// Retrieve list of users from the database with UserType as "Applicant"
	cursor, err := collection.Find(context.Background(), bson.M{"usertype": "Applicant"})
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var users []model.User
	for cursor.Next(context.Background()) {
		var user model.User
		err := cursor.Decode(&user)
		if err != nil {
			http.Error(w, "Failed to decode user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Fetching extracted data of an applicant
func GetApplicant(w http.ResponseWriter, r *http.Request) {
	applicantID := r.URL.Query().Get("applicant_id")

	if applicantID == "" {
		http.Error(w, "Applicant ID is required", http.StatusBadRequest)
		return
	}

	userType := r.Context().Value("userType").(string)

	if userType != "Admin" {
		http.Error(w, "Only Admin type users can access this API", http.StatusUnauthorized)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(applicantID)
	if err != nil {
		http.Error(w, "Invalid applicant ID", http.StatusBadRequest)
		return
	}

	var applicant model.Applicant
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&applicant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Applicant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch applicant data", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applicant)
}

// Fetching job openings
func GetJobs(w http.ResponseWriter, r *http.Request) {
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var jobs []model.Job
	for cursor.Next(context.Background()) {
		var job model.Job
		err := cursor.Decode(&job)
		if err != nil {
			http.Error(w, "Failed to decode job", http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
	}

	data, err := json.Marshal(jobs)
	if err != nil {
		http.Error(w, "Failed to marshal job data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Applying to a particular job
func ApplyForJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	if jobID == "" {
		http.Error(w, "Job ID not provided", http.StatusBadRequest)
		return
	}

	var job model.Job
	err := collection.FindOne(context.Background(), bson.M{"_id": jobID}).Decode(&job)
	if err != nil {
		http.Error(w, "Failed to find job", http.StatusNotFound)
		return
	}

	response := map[string]string{"message": "Applied for job successfully"}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
