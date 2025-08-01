package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Booking struct {
	ID             string `json:"id"`              // booking UUID
	SeatNumber     string `json:"seat_number"`     // e.g. “12A”
	BookingDate    string `json:"booking_date"`    // ISO 8601, e.g. “2025-08-01”
	FoodPreference string `json:"food_preference"` // “veg” | “non-veg”
}

// In memory store
var store = struct {
	sync.RWMutex
	data map[string]*Booking
}{data: make(map[string]*Booking)}

// POST /book defined to book airline ticket
func bookHandler(w http.ResponseWriter, r *http.Request) {
	var b Booking
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest); return
	}
	store.Lock(); defer store.Unlock()
	if _, ok := store.data[b.ID]; ok {
		http.Error(w, "booking exists", http.StatusConflict); return
	}
	store.data[b.ID] = &b
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(b)
}

// PUT /seat/{id} defined to book seat number of plane
func seatHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/seat/"):]
	var req struct{ Seat string `json:"seat_number"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest); return
	}
	store.Lock(); defer store.Unlock()
	if b, ok := store.data[id]; ok {
		b.SeatNumber = req.Seat; _ = json.NewEncoder(w).Encode(b)
		return
	}
	http.NotFound(w, r)
}

// PUT /date/{id} defined to update date of travel
func dateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/date/"):]
	var req struct{ Date string `json:"booking_date"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest); return
	}
	store.Lock(); defer store.Unlock()
	if b, ok := store.data[id]; ok {
		b.BookingDate = req.Date; _ = json.NewEncoder(w).Encode(b)
		return
	}
	http.NotFound(w, r)
}

// PUT /meal/{id} add meal preference
func mealHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/meal/"):]
	var req struct{ Meal string `json:"food_preference"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest); return
	}
	if req.Meal != "veg" && req.Meal != "non-veg" {
		http.Error(w, "meal must be veg|non-veg", http.StatusBadRequest); return
	}
	store.Lock(); defer store.Unlock()
	if b, ok := store.data[id]; ok {
		b.FoodPreference = req.Meal; _ = json.NewEncoder(w).Encode(b)
		return
	}
	http.NotFound(w, r)
}

// Defines APIs supported by the microservice
func main() {
	http.HandleFunc("/book", bookHandler)
	http.HandleFunc("/seat/", seatHandler)
	http.HandleFunc("/date/", dateHandler)
	http.HandleFunc("/meal/", mealHandler)

	log.Println("booking-svc listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
