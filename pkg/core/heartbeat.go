package core

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func BeginHeartbeatLoop(srv *Server) {
	log.Printf("Starting ClassiCube.net Heartbeat...")
	firstSent := false
	for {
		v := url.Values{}
		v.Set("name", srv.name)
		v.Set("port", srv.port)
		v.Set("users", strconv.Itoa(len(srv.players)))
		v.Set("max", strconv.Itoa(int(srv.maxUsers)))
		v.Set("public", strconv.FormatBool(srv.public))
		v.Set("salt", srv.Salt)
		v.Set("software", "Midnight")
		v.Set("web", "false")

		res, err := http.PostForm("http://www.classicube.net/server/heartbeat/", v)

		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
			break
		}

		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()

		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
			break
		}

		if !firstSent {
			log.Printf("Heartbeat sent: %s", data)
			firstSent = true
		}

		time.Sleep(time.Minute)
	}
}
