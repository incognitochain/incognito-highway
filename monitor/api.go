package monitor

import (
	"encoding/json"
	"fmt"
	"highway/process/simulateutils"
	"io/ioutil"
	"net/http"
)

// type

type ScenesAPI struct {
	Scenario *simulateutils.Scenario
}

func (s *ScenesAPI) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	switch r.Method {
	case http.MethodPost:
		var newScene map[string]simulateutils.Scene
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
		}
		err = json.Unmarshal(reqBody, &newScene)
		if err != nil {
			fmt.Fprintf(w, "Unmarshal requestBody failed, err %v", err)
		} else {
			err = s.Scenario.SetScenes(newScene)
			if err != nil {
				fmt.Fprintf(w, "Set scenes failed, err %v", err)
			}
		}
	case http.MethodGet:
		js, err := json.Marshal(*s.Scenario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
