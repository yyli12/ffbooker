package booker

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func arrayToString(a []int, delim string) string {
	return strings.Replace(fmt.Sprint(a), " ", delim, -1)
}

func SearchClass(programIDs []int, clubFilter *ClubFilter, fromTs int64, pageID int) ([]*Class, error) {
	token, tokenErr := GetToken()
	if tokenErr != nil {
		return nil, tokenErr
	}

	params := url.Values{
		"pageNumber": {fmt.Sprintf("%d", pageID)},
		"pageSize":   {"50"},
		"minPrice":   {"0.0"},
		"maxPrice":   {"150.0"},
		"programIds": {arrayToString(programIDs, ",")},
		"fromDate":   {fmt.Sprintf("%d", fromTs)},
	}
	if clubFilter.ClubID > 0 {
		params["clubId"] = []string{fmt.Sprintf("%d", clubFilter.ClubID)}
	} else {
		params["latitude"] = []string{fmt.Sprintf("%f", clubFilter.Latitude)}
		params["longitude"] = []string{fmt.Sprintf("%f", clubFilter.Longitude)}
		params["distance"] = []string{fmt.Sprintf("%f", clubFilter.Distance)}
	}

	client := &http.Client{}
	searchUrl := "https://api-mobile.circuithq.com/api/v2/class/search/?" + params.Encode()
	req, clientErr := http.NewRequest("GET", searchUrl, nil)
	if clientErr != nil {
		return nil, clientErr
	}
	req = addHeaders(req, token)

	respStruct := &struct {
		Error *APIError `json:"error"`
		Data  []*Class  `json:"data"`
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

func GetClass(classID int) (*Class, error) {
	token, tokenErr := GetToken()
	if tokenErr != nil {
		return nil, tokenErr
	}

	client := &http.Client{}
	getClassUrl := fmt.Sprintf("https://api-mobile.circuithq.com/api/v2/class/%d", classID)
	req, clientErr := http.NewRequest("GET", getClassUrl, nil)
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
	return respStruct.Data, nil
}
