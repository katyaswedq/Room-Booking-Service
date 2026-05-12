package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	baseURL    = "http://localhost:8080"
	roomsCount = 50
)

func main() {
	client := &http.Client{}

	token := mustGetAdminToken(client)

	for i := 1; i <= roomsCount; i++ {
		roomID := mustCreateRoom(client, token, i)
		mustCreateSchedule(client, token, roomID)
		fmt.Printf("created room %d: %s\n", i, roomID)
	}

	fmt.Println("seeding complete")
}

func mustGetAdminToken(client *http.Client) string {
	resp, err := client.Post(
		baseURL+"/dummyLogin",
		"application/json",
		bytes.NewBufferString(`{"role":"admin"}`),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("dummyLogin failed with status %d", resp.StatusCode)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		log.Fatal(err)
	}
	if tokenResp.Token == "" {
		log.Fatal("empty token from dummyLogin")
	}

	return tokenResp.Token
}

func mustCreateRoom(client *http.Client, token string, i int) string {
	body := fmt.Sprintf(`{"name":"Room %d","description":"Bench room %d","capacity":10}`, i, i)

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/rooms/create",
		bytes.NewBufferString(body),
	)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("create room failed with status %d", resp.StatusCode)
	}

	var roomResp struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&roomResp); err != nil {
		log.Fatal(err)
	}
	if roomResp.Room.ID == "" {
		log.Fatal("empty room id")
	}

	return roomResp.Room.ID
}

func mustCreateSchedule(client *http.Client, token, roomID string) {
	body := `{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/rooms/%s/schedule/create", baseURL, roomID),
		bytes.NewBufferString(body),
	)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("create schedule failed with status %d for room %s", resp.StatusCode, roomID)
	}
}