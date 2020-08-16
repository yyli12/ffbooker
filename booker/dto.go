package booker

import (
	"fmt"
	"time"
)

type BookingMode int

const (
	NonConcurrent BookingMode = iota
	Concurrent
)

type APIError struct {
	Code     int `json:"code"`
	Messages []*struct {
		Message string `json:"message"`
	} `json:"messages"`
}

func (e APIError) Error() string {
	errorMessage := ""
	for i, m := range e.Messages {
		if i == 0 {
			errorMessage = m.Message
			continue
		}
		errorMessage = errorMessage + ", " + m.Message
	}
	return fmt.Sprintf("%s (code: %d)", errorMessage, e.Code)
}

type ClubFilter struct {
	ClubID    int
	Latitude  float64
	Longitude float64
	Distance  float64
}

type Club struct {
	ID          int    `json:"clubId"`
	Name        string `json:"name"`
	DisplayName string `json:"clubWebsiteName"`
}

func (club Club) String() string {
	return fmt.Sprintf("[ID: %2d] %s", club.ID, club.DisplayName)
}

type Booking struct {
	ID      int   `json:"id"`
	StartAt int64 `json:"start_at"`
}

type Class struct {
	ID        int      `json:"classId"`
	Club      *Club    `json:"club"`
	Name      string   `json:"name"`
	Capacity  int64    `json:"capacity"`
	TimeStart int64    `json:"timeStart"`
	TimeEnd   int64    `json:"timeEnd"`
	Booking   *Booking `json:"booking"`
}

func (class Class) String() string {
	return fmt.Sprintf("[%s (Remaining %d)] %s at %s", class.Name, class.Capacity, class.Club.DisplayName, time.Unix(class.TimeStart, 0).Format("Mon, 02 Jan 2006 15:04:05"))
}
