package booker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/yyli12/ffbooker/log"
)

const (
	maxSearchPage                = 3
	maxTryBookingTime            = 10000
	defaultSearchClassInterval   = 300 * time.Millisecond
	defaultBookingInterval       = 1 * time.Second
	regularCheckIntervalMin      = 2 * time.Second
	regularCheckIntervalDelta    = 3 * time.Second
	concurrentBookingThreshold   = 5 * time.Second
	longBookingIntervalThreshold = 2 * time.Minute
	shortBookingInterval         = 20 * time.Second
)

func BookGymFloor(clubID int, slotTime *time.Time) error {
	slotTimeUnix := slotTime.Unix()
	searchPage := 1
	class := (*Class)(nil)
	for class == nil && searchPage <= maxSearchPage {
		classes, searchErr := SearchClass(ProgramIDsGymFloor, &ClubFilter{ClubID: clubID}, slotTimeUnix-60, searchPage)
		if searchErr != nil {
			log.Error("search class error: %s", searchErr)
			return searchErr
		}
		for _, c := range classes {
			if c.TimeStart == slotTimeUnix {
				class = c
				break
			}
		}
		searchPage++
		time.Sleep(defaultSearchClassInterval)
	}
	if class == nil {
		log.Error("no wanted class found in first %d pages", maxSearchPage)
		return fmt.Errorf("no wanted class found")
	}

	booking := (*Booking)(nil)
	getClassErr := error(nil)
	tryBookingTime := 1
	for booking == nil && tryBookingTime < maxTryBookingTime {
		interval := defaultBookingInterval
		bookingMode := NonConcurrent

		log.Info("trying %d time", tryBookingTime)
		class, getClassErr = GetClass(class.ID)
		if getClassErr != nil {
			log.Error("get class error: %s", getClassErr)
			return getClassErr
		}
		if class.Booking != nil {
			booking = class.Booking
			continue
		}
		if class.TimeStart < time.Now().Unix() {
			log.Error("class started")
			break
		}

		// capacity > 0: can try to book
		if class.Capacity > 0 {
			bookingError := error(nil)
			booking, bookingError = bookGymFloor(class, bookingMode)
			if booking != nil {
				break
			}
			if bookingError != nil {
				if apiError, ok := bookingError.(*APIError); ok && apiError != nil {
					if len(apiError.Messages) > 0 {
						message := apiError.Messages[0].Message
						if message == ErrorBookingTooSoon {
							timeToConcurrentBook := time.Unix(class.TimeStart, 0).Add(-46*time.Hour - concurrentBookingThreshold)
							stillHowLongToBook := timeToConcurrentBook.Sub(time.Now())
							if stillHowLongToBook <= 0 {
								// can start book soon! on fire!!!
								log.Warn("concurrent booking %s", log.TimeFormat(timeToConcurrentBook))
								interval = 100 * time.Millisecond
								bookingMode = Concurrent
							} else {
								log.Warn("too early to book, now %s, can only book after %s, %s then start to book", log.TimeFormat(time.Now()), log.TimeFormat(timeToConcurrentBook), stillHowLongToBook)
								if stillHowLongToBook > longBookingIntervalThreshold {
									interval = stillHowLongToBook - longBookingIntervalThreshold + 1*time.Second
								} else {
									interval = shortBookingInterval
								}

								if interval > stillHowLongToBook {
									interval = stillHowLongToBook + 1*time.Second
								}
							}
						}
					}
				}
			} else {
				log.Error("booking error", bookingError)
			}
		} else {
			// low want fish - try to book canceled slots
			regularCheckInterval := regularCheckIntervalMin + time.Duration(rand.Int63n(int64(regularCheckIntervalDelta)))
			log.Warn("no slot, regular booking every %s", regularCheckInterval)
			interval = regularCheckInterval
		}
		tryBookingTime++
		log.Warn("sleep %s", interval)
		time.Sleep(interval)
	}
	if booking == nil {
		log.Error("fail to book %s after %d attempts", class, maxTryBookingTime)
		return fmt.Errorf("fail to book")
	}

	log.Info("book %s successfully", class)
	return nil
}

// todo make it booking many times in parallel
func bookGymFloor(class *Class, mode BookingMode) (*Booking, error) {
	if class == nil {
		return nil, fmt.Errorf("class is nil")
	}

	if class.Booking != nil {
		return class.Booking, fmt.Errorf("%s already booked", class)
	}

	if class.Capacity <= 0 {
		return nil, fmt.Errorf("%s has no capacity", class)
	}

	token, tokenErr := GetToken()
	if tokenErr != nil {
		return nil, tokenErr
	}

	client := &http.Client{}
	bookingUrl := "https://api-mobile.circuithq.com/api/v2/class/book"
	reqBody := map[string]string{
		"class_id": fmt.Sprintf("%d", class.ID),
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, clientErr := http.NewRequest("POST", bookingUrl, bytes.NewBuffer(reqBodyBytes))
	if clientErr != nil {
		return nil, clientErr
	}
	req = addHeaders(req, token)

	respStruct := &struct {
		Error *APIError `json:"error"`
		Data  *Class    `json:"data"`
	}{}
	requestErr := sendRequest(client, req, respStruct)
	if requestErr != nil {
		return nil, requestErr
	}
	if respStruct.Error != nil {
		return nil, respStruct.Error
	}
	return respStruct.Data.Booking, nil
}
