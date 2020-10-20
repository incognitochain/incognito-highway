package monitor

import (
	"encoding/json"
	"fmt"
	"highway/process/simulateutils"
	"io/ioutil"
	"net/http"
)

type ForkAPI struct {
	FMaker simulateutils.ForkMaker
}

type SceneInput struct {
	Type   simulateutils.SceneType `json:"Type"`
	Scenes interface{}             `json:"Scene"`
	CID    int                     `json:"CID"`
}

func (f *ForkAPI) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if f.FMaker == nil {
		fmt.Fprintf(w, "ForkMaker is nil")
	} else {
		switch r.Method {
		case http.MethodPost:
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
			}
			newInput := &SceneInput{}
			err = json.Unmarshal(reqBody, &newInput)
			if err != nil {
				fmt.Fprintf(w, "Unmarshal requestBody failed, err %v", err)
				return
			}
			x, err := json.Marshal(newInput.Scenes)
			if err != nil {
				fmt.Fprintf(w, "Marshal scenes failed, err %v", err)
				return
			}
			err = f.FMaker.SetScenes(newInput.Type, x, newInput.CID)

			if err != nil {
				fmt.Fprintf(w, "SetScenes failed, err %v", err)
			}
			// case http.MethodGet:
			// 	js, err := json.Marshal(*s.Scenario)
			// 	if err != nil {
			// 		http.Error(w, err.Error(), http.StatusInternalServerError)
			// 		return
			// 	}

			// 	w.Header().Set("Content-Type", "application/json")
			// 	w.Write(js)
		}
	}
}
