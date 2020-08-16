package main

import (
	"github.com/yyli12/ffbooker/booker"
)

func main() {
	if booker.EnsureToken() != nil {
		return
	}

	club, selectClubErr := booker.SelectClub()
	if selectClubErr != nil {
		return
	}

	time, inputTimeErr := booker.InputTime()
	if inputTimeErr != nil {
		return
	}

	if booker.BookGymFloor(club.ID, time) != nil {
		return
	}
}
