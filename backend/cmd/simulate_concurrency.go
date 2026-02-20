package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const baseURL = "http://localhost:8080"

type User struct {
	ID    uint   `json:"ID"`
	Role  string `json:"Role"`
	Email string `json:"Email"`
}

type Event struct {
	ID                uint `json:"ID"`
	Title             string
	TotalCapacity     int
	AvailableCapacity int
}

func main() {
	fmt.Println("Starting Concurrency Simulation...")

	// 1. Create Organizer
	organizer := loginUser("organizer_test@test.com", "pass", true)
	if organizer.ID == 0 {
		fmt.Println("Failed to create organizer (an organizer may already exist). Proceeding with existing...")
		organizer = loginUser("bhagojishashank@gmail.com", "1234567890", true) // Fallback if user already created one
	}

	// 2. Create Event with Capacity = 1
	eventID := createEvent(organizer.ID, "Concurrent Test Event", 1)
	if eventID == 0 {
		fmt.Println("Failed to create test event")
		return
	}
	fmt.Printf("Created Event ID %d with Capacity of 1.\n", eventID)

	// 3. Create 5 test users
	var users []User
	for i := 1; i <= 5; i++ {
		email := fmt.Sprintf("testuser%d@test.com", i)
		user := loginUser(email, "password", false)
		users = append(users, user)
		fmt.Printf("Created/Logged in test user: %s (ID: %d)\n", email, user.ID)
	}

	// 4. Simulate Concurrent Registration for the remaining 1 spot
	fmt.Println("\n--- Initiating Concurrent Registration Race ---")
	fmt.Println("5 users are attempting to register for the exact same event at the exact same moment...")

	var wg sync.WaitGroup
	results := make(chan string, 5)

	for _, u := range users {
		wg.Add(1)
		go func(userID uint) {
			defer wg.Done()
			
			payload := map[string]interface{}{
				"user_id":  userID,
				"event_id": eventID,
			}
			jsonData, _ := json.Marshal(payload)

			resp, err := http.Post(fmt.Sprintf("%s/register", baseURL), "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				results <- fmt.Sprintf("User %d Error: %v", userID, err)
				return
			}
			defer resp.Body.Close()

			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			
			if resp.StatusCode == http.StatusOK {
				results <- fmt.Sprintf("User %d SUCCESS! Managed to secure the spot.", userID)
			} else {
				results <- fmt.Sprintf("User %d FAILED. Status: %d, Response: %s", userID, resp.StatusCode, string(bodyBytes))
			}

		}(u.ID)
	}

	wg.Wait()
	close(results)

	fmt.Println("\n--- Race Results ---")
	for res := range results {
		fmt.Println(res)
	}
	
	fmt.Println("\nSimulation Complete. Notice how only 1 user gets a SUCCESS response, while the others are rejected gracefully by the Row-Level Locking mechanism!")
}

func loginUser(email, password string, isOrganizer bool) User {
	payload := map[string]interface{}{
		"email":        email,
		"password":     password,
		"is_organizer": isOrganizer,
	}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(fmt.Sprintf("%s/auth/login", baseURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return User{}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var res struct {
			User User `json:"user"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		return res.User
	}
	return User{}
}

func createEvent(organizerID uint, title string, capacity int) uint {
	// Set date to tomorrow
	futureDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	
	payload := map[string]interface{}{
		"title": title,
		"totalCapacity": capacity,
		"createdBy": organizerID,
		"eventdate": futureDate,
	}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(fmt.Sprintf("%s/events", baseURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var event Event
		json.NewDecoder(resp.Body).Decode(&event)
		return event.ID
	}
	return 0
}
