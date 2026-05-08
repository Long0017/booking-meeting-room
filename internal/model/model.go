package model

import "time"

type Room struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	Location  string    `json:"location"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Booking struct {
	ID             int       `json:"id"`
	RoomID         int       `json:"room_id"`
	RoomName       string    `json:"room_name,omitempty"`
	EmployeeID     string    `json:"employee_id"`
	Department     string    `json:"department"`
	Title          string    `json:"title"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	AttendeesCount int       `json:"attendees_count"`
	Notes          string    `json:"notes"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateBookingRequest struct {
	RoomID         int    `json:"room_id"`
	EmployeeID     string `json:"employee_id"`
	Department     string `json:"department"`
	Title          string `json:"title"`
	StartTime      string `json:"start_time"` // RFC3339
	EndTime        string `json:"end_time"`   // RFC3339
	AttendeesCount int    `json:"attendees_count"`
	Notes          string `json:"notes"`
	CancelCode     string `json:"cancel_code"`
}

type CancelBookingRequest struct {
	CancelCode string `json:"cancel_code"`
}
