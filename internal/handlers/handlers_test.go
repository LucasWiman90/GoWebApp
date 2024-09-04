package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/LucasWiman90/GoWebApp/internal/driver"
	"github.com/LucasWiman90/GoWebApp/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"nebula", "/nebula", "GET", http.StatusOK},
	{"darkfathom", "/darkfathom", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non-existent", "/mystical-house", "GET", http.StatusNotFound},
	//new routes
	{"login", "/user/login", "GET", http.StatusOK},
	{"login", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Nebula Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: god %d, wanted %d", rr.Code, http.StatusOK)
	}

	//test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: god %d, wanted %d", rr.Code, http.StatusOK)
	}

	//test case with non-existant room, roomID that doesnt exist
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: god %d, wanted %d", rr.Code, http.StatusOK)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	/*layout := "2006-01-02"
	sd, _ := time.Parse(layout, "2024-01-02")
	ed, _ := time.Parse(layout, "2024-01-03")

	// Base reservation used across test cases
	baseReservation := models.Reservation{
		RoomID:    1,
		StartDate: sd,
		EndDate:   ed,
		Room: models.Room{
			ID:       1,
			RoomName: "Nebula Quarters",
		},
	} */

	// Define a slice of test cases
	tests := []struct {
		name             string
		postedData       url.Values
		isNilBody        bool
		expectedStatus   int
		expectedLocation string
		expectedHTML     string
	}{
		{
			name: "Valid_reservation",
			postedData: url.Values{
				"start_date": {"2050-01-01"},
				"end_date":   {"2050-01-02"},
				"first_name": {"Harold"},
				"last_name":  {"Jones"},
				"email":      {"harold.jones@gmail.com"},
				"phone":      {"9718594945"},
				"room_id":    {"1"}},
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/reservation-summary",
		},
		{
			name:             "Missing_post_body",
			postedData:       nil,
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/",
		},
		{
			name: "Invalid_start_date",
			postedData: url.Values{
				"start_date": {"invalid"},
				"end_date":   {"2050-01-02"},
				"first_name": {"Wilburt"},
				"last_name":  {"Sonesson"},
				"email":      {"wilburt@sonesson.com"},
				"phone":      {"555-555-5555"},
				"room_id":    {"1"},
			},
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/",
		},
		{
			name: "Invalid_end_date",
			postedData: url.Values{
				"start_date": {"2050-01-02"},
				"end_date":   {"invalid"},
				"first_name": {"Wilburt"},
				"last_name":  {"Sonesson"},
				"email":      {"wilburt@sonesson.com"},
				"phone":      {"555-555-5555"},
				"room_id":    {"1"},
			},
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/",
		},
		{
			name: "Invalid_room_id",
			postedData: url.Values{
				"start_date": {"2050-01-01"},
				"end_date":   {"2050-01-02"},
				"first_name": {"Ronald"},
				"last_name":  {"McDonald"},
				"email":      {"ronald@mcdonald.com"},
				"phone":      {"666-444-5555"},
				"room_id":    {"invalid"},
			},
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/",
		},
		{
			name: "Invalid_data_(first_name)",
			postedData: url.Values{
				"start_date": {"2050-01-01"},
				"end_date":   {"2050-01-02"},
				"first_name": {"5"},
				"last_name":  {"Numberman"},
				"email":      {"x@Numberman.com"},
				"phone":      {"555-555-5555"},
				"room_id":    {"1"},
			},
			expectedStatus:   http.StatusOK,
			expectedHTML:     `action="/make-reservation"`,
			expectedLocation: "",
		},
		{
			name: "DB_Insert_fails_reservation",
			postedData: url.Values{
				"start_date": {"2050-01-01"},
				"end_date":   {"2050-01-02"},
				"first_name": {"Benjamin"},
				"last_name":  {"Numberman"},
				"email":      {"Benjamin@Numberman.com"},
				"phone":      {"111-222-3333"},
				"room_id":    {"2"},
			},
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/",
		},
		{
			name: "DB_Insert_fails_restriction",
			postedData: url.Values{
				"start_date": {"2050-01-01"},
				"end_date":   {"2050-01-02"},
				"first_name": {"Benjamin"},
				"last_name":  {"Numberman"},
				"email":      {"Benjamin@Numberman.com"},
				"phone":      {"111-222-3333"},
				"room_id":    {"1000"},
			},
			expectedStatus:   http.StatusSeeOther,
			expectedHTML:     "",
			expectedLocation: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.postedData != nil {
				req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(tt.postedData.Encode()))
			} else {
				req, _ = http.NewRequest("POST", "/make-reservation", nil)
			}

			// Set up context and session if needed
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(Repo.PostReservation)

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("%s handler returned wrong response code: got %d, wanted %d", tt.name, rr.Code, tt.expectedStatus)
			}

			if tt.expectedLocation != "" {
				// get the URL from test
				actualLoc, _ := rr.Result().Location()
				if actualLoc.String() != tt.expectedLocation {
					t.Errorf("failed %s: expected location %s, but got location %s", tt.name, tt.expectedLocation, actualLoc.String())
				}
			}

			if tt.expectedHTML != "" {
				// read the response body into a string
				html := rr.Body.String()
				if !strings.Contains(html, tt.expectedHTML) {
					t.Errorf("failed %s: expected to find %s but did not", tt.name, tt.expectedHTML)
				}
			}
		})
	}
}

func TestRepositry_ReservationSummary(t *testing.T) {
	//Create a reservation
	layout := "2006-01-02"
	sd, _ := time.Parse(layout, "2024-01-02")
	ed, _ := time.Parse(layout, "2024-01-03")
	reservation := models.Reservation{
		RoomID:    1,
		StartDate: sd,
		EndDate:   ed,
		Room: models.Room{
			ID:       1,
			RoomName: "Nebula Quarters",
		},
	}

	// First case -- reservation in session
	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// Put the reservation in the session
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// Second case -- reservation not in session
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}
}

func TestRepositry_ChooseRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Nebula Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	session.Put(ctx, "reservation", reservation)
	req.RequestURI = "/choose-room/1"
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// Case 2 when reservation not in session
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// Case 3 when atoi fails
	req, _ = http.NewRequest("GET", "/choose-room/notFound", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	session.Put(ctx, "reservation", reservation)

	req.RequestURI = "/choose-room/notFound"
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

}

func TestRepositry_BookRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Nebula Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/book-room?s=2050-01-01&e=20250-01-02&id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// Case 2 database failed
	req, _ = http.NewRequest("GET", "/book-room?s=2040-01-01&e=2040-01-02&id=99", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	// Case 1 - Room not available
	postedData := url.Values{}
	postedData.Add("start", "2050-01-02")
	postedData.Add("end", "2050-01-03")
	postedData.Add("room_id", "1")

	// create our request
	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder
	rr := httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	// make the request to our handler
	handler.ServeHTTP(rr, req)

	// since we have no rooms available, we expect to get status http.StatusSeeOther
	// this time we want to parse JSON and get the expected response
	var j jsonResponse
	err := json.Unmarshal([]byte(rr.Body.Bytes()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date > 2049-12-31, we expect no availability
	if j.OK {
		t.Error("Got availability when none was expected in AvailabilityJSON")
	}

	// Case 2 Parsing failure - No Request Body
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal([]byte(rr.Body.Bytes()), &j)

	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if j.Message != "parse-fail:internal server error" {
		t.Error("It should be fail and its passed")
	}

	// Case 3 Room available
	postedData = url.Values{}
	postedData.Add("start", "2040-01-02")
	postedData.Add("end", "2040-01-03")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal([]byte(rr.Body.Bytes()), &j)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if !j.OK {
		t.Error("There is no availablity and it should be fail and its passed")
	}

	// Cas 4 Database error
	postedData = url.Values{}
	postedData.Add("start", "2060-01-01")
	postedData.Add("end", "2060-01-02")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal([]byte(rr.Body.Bytes()), &j)

	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if j.Message != "Error querying database" {
		t.Error("Error connecting to my DB:it should be fail and its passed")
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name           string
		reqBody        url.Values
		isNilBody      bool
		expectedStatus int
	}{
		{
			name:           "Empty post body",
			isNilBody:      true, // Indicates that the request body should be nil
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Invalid start date",
			reqBody:        url.Values{"start": {"invalid"}, "end": {"2040-01-02"}},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Invalid end date",
			reqBody:        url.Values{"start": {"2040-01-01"}, "end": {"invalid"}},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Database query fails",
			reqBody:        url.Values{"start": {"2060-01-01"}, "end": {"2060-01-02"}},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "No rooms available",
			reqBody:        url.Values{"start": {"2050-01-01"}, "end": {"2050-01-02"}},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "Rooms available",
			reqBody:        url.Values{"start": {"2024-01-01"}, "end": {"2024-01-02"}},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.isNilBody {
				// Handle the case where the request body should be nil
				req, _ = http.NewRequest("POST", "/search-availability", nil)
			} else {
				// Encode the url.Values into the request body
				req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(tt.reqBody.Encode()))
			}

			// Get the context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.PostAvailability)

			// Make the request to our handler
			handler.ServeHTTP(rr, req)

			// Check if the status code is what we expect
			if rr.Code != tt.expectedStatus {
				t.Errorf("%s gave wrong status code: got %d, wanted %d", tt.name, rr.Code, tt.expectedStatus)
			}
		})
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"me@here.se",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"john@scatman.com",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid-data",
		"x",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

func TestLogin(t *testing.T) {
	//range through all tests
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		//Create request
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		//Set header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		//Call the handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			//Get the URL from the test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		//Checking for expected values in html
		if e.expectedHTML != "" {
			//Read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but did not", e.name, e.expectedHTML)
			}
		}
	}
}

func TestAdminDeleteReservation(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name                 string
		queryParams          string
		expectedResponseCode int
		expectedLocation     string
	}{
		{
			name:                 "delete-reservation",
			queryParams:          "",
			expectedResponseCode: http.StatusSeeOther,
			expectedLocation:     "",
		},
		{
			name:                 "delete-reservation-back-to-calender",
			queryParams:          "?y=2024&m=12",
			expectedResponseCode: http.StatusSeeOther,
			expectedLocation:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", tt.queryParams), nil)
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.AdminDeleteReservation)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Errorf("failed %s: expected code %d, but got %d", tt.name, tt.expectedResponseCode, rr.Code)
			}
		})
	}
}

func TestAdminProcessReservation(t *testing.T) {
	// Define a slice of test cases
	tests := []struct {
		name                 string
		queryParams          string
		expectedResponseCode int
		expectedLocation     string
	}{
		{
			name:                 "process-reservation",
			queryParams:          "",
			expectedResponseCode: http.StatusSeeOther,
			expectedLocation:     "",
		},
		{
			name:                 "process-reservation-back-to-calender",
			queryParams:          "?y=2024&m=12",
			expectedResponseCode: http.StatusSeeOther,
			expectedLocation:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", tt.queryParams), nil)
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.AdminProcessReservation)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Errorf("failed %s: expected code %d, but got %d", tt.name, tt.expectedResponseCode, rr.Code)
			}
		})
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
