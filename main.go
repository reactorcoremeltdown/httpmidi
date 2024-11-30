package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

type Payload struct {
	Kind    string `json:"kind"`
	Id      uint8  `json:"id"`
	Value   uint8  `json:"value"`
	Channel uint8  `json:"channel"`
}

func Route(out drivers.Out) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			res.WriteHeader(400)
			fmt.Fprint(res, "Unable to parse forms")
		} else {
			payload := req.Form.Get("payload")
			var inc Payload
			json.Unmarshal([]byte(payload), &inc)

			switch inc.Kind {
			case "NoteOn":
				out.Send(midi.NoteOn(inc.Channel, inc.Id, inc.Value))
				fmt.Fprint(res, "OK")
			case "NoteOff":
				out.Send(midi.NoteOff(inc.Channel, inc.Id))
				fmt.Fprint(res, "OK")
			default:
				res.WriteHeader(404)
				fmt.Fprint(res, "Not implemented")
			}

		}

	}
}

func main() {
	driver, err := rtmididrv.New()
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	outPort, err := driver.OpenVirtualOut("httpmidi_out")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/input", Route(outPort))
	http.Handle("/", r)

	socket := "/tmp/httpmidi.sock"

	conn, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Fatal(http.Serve(conn, r))
}
