package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	automate()
}

func automate() {
	for {
		gamesChan := make(chan [10][5]interface{})
		standingsChan := make(chan [20][4]interface{})
		roundIdChan := make(chan int)
		var s string = "no games this week"
		var sleepTime int = 120

		go getGames(gamesChan)
		go getStandings(standingsChan)
		go getRoundId(roundIdChan)

		roundId := <-roundIdChan

		homeTeam, awayTeam, odds, parentMatchId, startTime, leading := extractGames(<-gamesChan, <-standingsChan)

		close(roundIdChan)
		close(gamesChan)
		close(standingsChan)

		if leading == 0 {
			log.Println(s)

		} else {
			if roundId > 0 {
				log.Println("Round Id: ", roundId)
				placeBet(
					homeTeam,
					awayTeam,
					odds,
					parentMatchId,
					startTime,
					leading,
				)
			} else {
				log.Println(s)
			}

		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Println("\n ")
		log.Fatal("AN ERROR OCCURED!!: ", err)
	}
}

func handleGetGamesResponse(response *http.Response) [10][5]interface{} {
	// Reading the response in bytes
	responseBody, err := ioutil.ReadAll(response.Body)
	handleError(err)

	// Declaring our struct type where the response data will be unmarshalled.
	// Its called 'BigData' because a very big json object will be unmarshalled into it.
	type BigData struct {
		Status_code        float64
		Status_description string
		Data               map[string]interface{}
	}

	// Initializing/ Instanciating the struct BigData
	m := BigData{}

	// Unmarshalling the response data into the struct variable
	err = json.Unmarshal(responseBody, &m)

	handleError(err)

	dat := m.Data["matches"].([]interface{})

	var games [10][5]interface{}

	for i := range dat {
		var data [5]interface{}
		_dat := dat[i].(map[string]interface{})
		data[0] = _dat["home_team"]
		data[1] = _dat["away_team"]
		data[2] = _dat["outcomes"]
		data[3] = _dat["parent_match_id"]
		data[4] = _dat["start_time"]

		games[i] = data
	}

	log.Println("Games and odds obtained.")
	return games

}

func handleGetStandingsResponse(response *http.Response) [20][4]interface{} {

	// Reading the response in bytes
	responseBody, err := ioutil.ReadAll(response.Body)
	handleError(err)

	// Declaring our struct type where the response data will be unmarshalled.
	// Its called 'BigData' because a very big json object will be unmarshalled into it.
	type BigData struct {
		Status_code        float64
		Status_description string
		Data               map[string]interface{}
	}

	// Initializing/ Instanciating the struct BigData
	m := BigData{}

	// Unmarshalling the response data into the struct variable
	err = json.Unmarshal(responseBody, &m)

	handleError(err)

	var standings [20][4]interface{}

	dat := m.Data["standings"].([]interface{})
	for i := range dat {
		var data [4]interface{}
		_dat := dat[i].(map[string]interface{})
		data[0] = _dat["team_id"]
		data[1] = _dat["team_name"]
		data[2] = _dat["team_form"]
		data[3] = _dat["points"]

		standings[i] = data
	}

	log.Println("Teams standings obtained.")
	return standings

}

func getRoundId(c chan int) {

	url := "https://odibets.com/api/fv"

	//Converting our map payload to json
	data, _ := json.Marshal(map[string]string{
		"competition_id": "1",
		"tab":            "results",
		"period":         "2021-12-24 11:55:00",
		"sub_type_id":    "",
	})

	// Converting payload to bytes
	payload := bytes.NewBuffer(data)

	// Our Http post request
	response, err := http.Post(url, "application/json", payload)
	handleError(err)

	// Close the connection after everything completes
	defer response.Body.Close()

	// Reading the response in bytes
	responseBody, err := ioutil.ReadAll(response.Body)
	handleError(err)

	// Declaring our struct type where the response data will be unmarshalled.
	// Its called 'BigData' because a very big json object will be unmarshalled into it.
	type BigData struct {
		Status_code        float64
		Status_description string
		Data               map[string]interface{}
	}

	// Initializing/ Instanciating the struct BigData
	m := BigData{}

	// Unmarshalling the response data into the struct variable
	err = json.Unmarshal(responseBody, &m)
	handleError(err)

	// This section of code simply checks which round it is. Rounds run from
	// 1 through 38. Data is fetched when the round is 38.
	dat := m.Data["results"].([]interface{})
	_dat := dat[0].(map[string]interface{})
	__dat := fmt.Sprint(_dat["round_id"])
	round_id, err := strconv.Atoi(__dat)
	handleError(err)

	log.Println("Round ID obtained.")
	c <- round_id
}

func getGames(c chan [10][5]interface{}) {
	url := "https://odibets.com/api/fv"

	//Converting our map payload to json
	data, _ := json.Marshal(map[string]string{
		"competition_id": "1",
		"tab":            "",
		"period":         "",
		"sub_type_id":    "",
	})

	// Converting payload to bytes
	payload := bytes.NewBuffer(data)

	// Our Http post request
	response, err := http.Post(url, "application/json", payload)
	handleError(err)

	games := handleGetGamesResponse(response)

	c <- games

}

func getStandings(c chan [20][4]interface{}) {
	url := "https://odibets.com/api/fv"

	//Converting our map payload to json
	data, _ := json.Marshal(map[string]string{
		"competition_id": "1",
		"tab":            "standings",
		"period":         "",
		"sub_type_id":    "",
	})

	// Converting payload to bytes
	payload := bytes.NewBuffer(data)

	// Our Http post request
	response, err := http.Post(url, "application/json", payload)
	handleError(err)

	standings := handleGetStandingsResponse(response)

	c <- standings

}

func extractGames(
	games [10][5]interface{},
	standings [20][4]interface{},
) (
	interface{},
	interface{},
	interface{},
	interface{},
	interface{},
	int,
) {

	if len(games[0][2].([]interface{})) < 1 {
		time.Sleep(45 * time.Second)
		return nil, nil, nil, nil, nil, 0

	} else {
		for _, match := range games {
			if match[0] == standings[0][1] && match[1] == standings[19][1] {
				if standings[0][0].(int) > standings[19][0].(int) && standings[0][0].(int) < 7 {
					return match[0], match[1], match[2], match[3], match[4], 1
				}

			} else if match[0] == standings[19][1] && match[1] == standings[0][1] {
				if standings[0][0].(int) > standings[19][0].(int) && standings[0][0].(int) < 7 {
					return match[0], match[1], match[2], match[3], match[4], 2
				}
			}
		}
	}
	return nil, nil, nil, nil, nil, 0
}

func placeBet(
	homeTeam interface{},
	awayTeam interface{},
	odds interface{},
	parentMatchId interface{},
	startTime interface{},
	leading int,
) {

	outcome_id := fmt.Sprint(leading)
	outcome_name := ""
	var odd_value interface{}

	if leading == 1 {
		outcome_name = "Home Team wins"
		odd_value = odds.([]interface{})[0].(map[string]interface{})["odd_value"]
	} else if leading == 2 {
		outcome_name = "Away Team wins"
		odd_value = odds.([]interface{})[2].(map[string]interface{})["odd_value"]
	}

	sub_type_id := "1X2"
	odd_type := "1X2"
	parent_match_id := parentMatchId
	home_team := homeTeam
	away_team := awayTeam
	start_time := startTime
	prev_odd_value := odd_value
	status := "1"
	live := 0
	sport_id := ""
	custom := fmt.Sprint(fmt.Sprint(parent_match_id) + odd_type + outcome_name)
	auto := 0

	var s [1]map[string]interface{}
	_s := map[string]interface{}{
		"outcome_id":      outcome_id,
		"outcome_name":    outcome_name,
		"sub_type_id":     sub_type_id,
		"odd_type":        odd_type,
		"parent_match_id": parent_match_id,
		"home_team":       home_team,
		"away_team":       away_team,
		"start_time":      start_time,
		"odd_value":       odd_value,
		"prev_odd_value":  prev_odd_value,
		"status":          status,
		"live":            live,
		"sport_id":        sport_id,
		"custom":          custom,
		"auto":            auto,
	}
	s[0] = _s

	type Payload struct {
		M      int64                     `json:"m"`
		Stake  int64                     `json:"stake"`
		S      [1]map[string]interface{} `json:"s"`
		Msisdn string                    `json:"msisdn"`
		Ref    string                    `json:"ref"`
		Pwd    string                    `json:"pwd"`
	}

	//Converting our map payload to json
	payload := Payload{
		M:      5,
		Stake:  10,
		S:      s,
		Msisdn: "",
		Ref:    "",
		Pwd:    "",
	}
	data, _ := json.Marshal(payload)

	placeBetUrl := "https://odibets.com/api/sb"

	request, err := http.NewRequest("POST", placeBetUrl, bytes.NewBuffer(data))
	handleError(err)
	request.Header.Set("Accept", "application/json, text/plain, */*")
	request.Header.Set("Accept-Encoding", "Accept-Encoding")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJUSEVfSVNTVUVSIiwiYXVkIjoiVEhFX0FVRElFTkNFIiwiaWF0IjoxNjQ3OTQwNzg3LCJuYmYiOjE2NDc5NDA3ODcsImRhdGEiOnsicHJvZmlsZV9pZCI6Ijc5MTkxNiIsIm1zaXNkbiI6IjI1NDcxNjkwOTkzNSIsInRoZW1lIjoiLTEiLCJzdGF0dXMiOjF9fQ.ClRuyejvxE6RFkFt8__OJuhrkjN7eNP2ClcKoxO4XyYaIn")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Cookie", "odibetskenya=1k8t622knal5uqp7g4qsl78ja0")
	request.Header.Set("Host", "odibets.com")
	request.Header.Set("Origin", "https://odibets.com")
	request.Header.Set("Referer", "https://odibets.com/")
	request.Header.Set("sec-ch-ua-platform", "Windows")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.74 Safari/537.36")

	client := &http.Client{}
	response, error := client.Do(request)
	handleError(error)

	log.Println("Placed bet successfully.")

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	response.Body.Close()

	// fmt.Print(string(data), "\n")

}
