package booker

import (
	"fmt"
	"net/http"
	"sort"
)

const (
	selectClubMaxTime = 5
)

func SelectClub() (*Club, error) {
	clubs, getClubsErr := getClubs()
	if getClubsErr != nil {
		return nil, getClubsErr
	}

	sort.Slice(clubs, func(i, j int) bool {
		return clubs[i].ID < clubs[j].ID
	})

	clubIDMap := map[int]*Club{}
	for _, club := range clubs {
		fmt.Println(club)
		clubIDMap[club.ID] = club
	}

	for i := 0; i < selectClubMaxTime; i++ {
		clubID := -1
		fmt.Println("Select Club ID: ")
		fmt.Scanf("%d", &clubID)
		if club, ok := clubIDMap[clubID]; ok {
			fmt.Printf("Selected %s\n", club)
			return club, nil
		} else {
			fmt.Println("Error Club ID Input")
		}
	}

	return nil, fmt.Errorf("fail to select club")
}

func getClubs() ([]*Club, error) {
	token, tokenErr := GetToken()
	if tokenErr != nil {
		return nil, tokenErr
	}

	client := &http.Client{}
	getClubUrl := "https://api-mobile.circuithq.com/api/v2/club/sg"
	req, clientErr := http.NewRequest("GET", getClubUrl, nil)
	if clientErr != nil {
		return nil, clientErr
	}
	req = addHeaders(req, token)

	respStruct := &struct {
		Error *APIError `json:"error"`
		Data  []*Club   `json:"data"`
	}{}
	requestErr := sendRequest(client, req, respStruct)
	if requestErr != nil {
		return nil, requestErr
	}
	if respStruct.Error != nil {
		return nil, respStruct.Error
	}
	return respStruct.Data, nil
}
