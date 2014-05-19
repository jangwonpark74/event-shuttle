package main

import (
	"github.com/bmizerany/pat"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
)

type Endpoint struct{
	store *Store
}

func StartEndpoint(port string, store *Store) *Endpoint {
	endpoint := Endpoint{store:store}
	mux := pat.New()
	mux.Post(fmt.Sprintf("/:channel"), http.HandlerFunc(endpoint.PostEvent))
	http.ListenAndServe("127.0.0.1:"+port, mux)
	return &endpoint
}

func (e *Endpoint)PostEvent(w http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get(":channel")
	if channel == "" {
		w.WriteHeader(400)
		w.Write(NoChannel)
	} else {
		body, err := ioutil.ReadAll(req.Body);
		if err != nil {
			w.WriteHeader(500)
			w.Write(BodyErr)
		} else {
			saved := make(chan bool)
			event := EventIn{channel:channel, body:body, saved: saved}
		    e.store.EventsInChannel() <- &event
			timeout := time.After(1 * time.Second)
			select {
			case ok := <-saved:
				if ok {
					w.WriteHeader(200)
				} else {
					w.WriteHeader(500)
					w.Write(SaveErr)
				}
			case <-timeout:
				w.WriteHeader(500)
				w.Write(SaveTimeout)
			}

		}
	}


}

var NoChannel = Json(ErrJson{id:"no-channel", message:"no event channel specified"})
var BodyErr = Json(ErrJson{id:"read-error", message:"error while reading body"})
var SaveErr = Json(ErrJson{id:"save-error", message:"save event returned false"})
var SaveTimeout = Json(ErrJson{id:"save-timeout", message:"save timed out"})

func Json(err ErrJson) []byte {
	j, _ := json.Marshal(err)
	return j
}

type ErrJson struct {
	id      string
	message string
}