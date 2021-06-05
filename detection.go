package main

import (
	"net"
	"strings"
	"time"

	mdns "github.com/hashicorp/mdns"
	log "github.com/sirupsen/logrus"
)

const (
	maxEntries = 20
)

func init() {
	// Disable mdns annoying and verbose logging
	//log.SetOutput(ioutil.Discard)
}

type PlugEntry struct {
	DetectionId string
	Id          string
	AddrV4      net.IP
	AddrV6      net.IP
}

// https://github.com/grasparv/go-chromecast/blob/master/dns/dns.go

func detectPlugs() []PlugEntry {
	log.Debug("Periodic plug detection ")
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, maxEntries)
	go func() {
		// This will find any and all google products, including chromecast, home mini, etc.
		mdns.Query(&mdns.QueryParam{
			Service: "_http._tcp",
			Domain:  "local",
			Timeout: time.Second * 3,
			Entries: entriesCh,
		})
		close(entriesCh)
	}()
	plugs := make([]PlugEntry, 0)
	for entry := range entriesCh {
		infoFields := make(map[string]string, len(entry.InfoFields))
		for _, infoField := range entry.InfoFields {
			splitField := strings.Split(infoField, "=")
			if len(splitField) != 2 {
				continue
			}
			infoFields[splitField[0]] = splitField[1]
		}
		if strings.HasPrefix(entry.Name, "shellyplug-s-") {
			plug := PlugEntry{
				DetectionId: infoFields["id"],
				AddrV4:      entry.AddrV4,
				AddrV6:      entry.AddrV6,
			}
			log.Debugf("PLUG http service found: %s %s", plug, entry)
			plugs = append(plugs, plug)
		}
	}

	return plugs
}
