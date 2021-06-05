package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// get plug description from /settings and /status
func get_plug_desc(plug_host string) (PlugDescription, error) {
	var plug_desc PlugDescription

	status, err := get_plug_status(plug_host)
	if err != nil {
		log.Println("Error getting response: ", err)
		return plug_desc, err
	}

	settings, err := get_plug_settings(plug_host)
	if err != nil {
		log.Println("Error getting response: ", err)
		return plug_desc, err
	}

	plug_desc = PlugDescription{
		Id:           settings.Device.Mac,
		Hostname:     settings.Device.Hostname,
		Name:         settings.Relays[0].Name,
		Type:         settings.Device.Type,
		LastSeen:     time.Now(),
		AddrV4:       status.Wifi_sta.Ip,
		Mac:          status.Mac,
		Is_available: true,
	}

	// log.Println("   POWER: ", m.Power)
	return plug_desc, nil
}

// type for /settings
type PlugSettings struct {
	Device struct {
		Mac      string
		Hostname string
		Type     string
	}
	Relays []struct {
		Name           string
		Appliance_type string
	}
}

// get plug settings from `http://<plug>/settings`
func get_plug_settings(plug_host string) (PlugSettings, error) {
	var plug_settings PlugSettings
	url := fmt.Sprintf("http://%v/settings", plug_host)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error getting response for /settings: ", plug_host, err)
		return plug_settings, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading settings response: ", err)
		return plug_settings, err
	}
	err = json.Unmarshal(body, &plug_settings)
	if err != nil {
		log.Println("ERROR parsing meter response: ", err)
		return plug_settings, err
	}
	return plug_settings, nil
}

// type for /status
type PlugStatus struct {
	Wifi_sta struct {
		Ssid string
		Ip   string
	}
	Mac string
}

// get plug status from `http://<plug>/status`
func get_plug_status(plug_host string) (PlugStatus, error) {
	var plug_status PlugStatus
	url := fmt.Sprintf("http://%v/status", plug_host)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error getting response for /status: ", plug_host, err)
		return plug_status, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading status response: ", err)
		return plug_status, err
	}
	err = json.Unmarshal(body, &plug_status)
	if err != nil {
		log.Println("ERROR parsing status response: ", err)
		return plug_status, err
	}
	return plug_status, nil
}

// type for storing the json response
// for GET requests <plug>/meter/0
type MeterInfo struct {
	Power     float64
	Overpower float64
	Is_valid  bool
	Timestamp uint64
	Total     uint32
	Counters  [3]float64
}

func get_energy_data(plug_host string) (MeterInfo, error) {
	var m MeterInfo

	// Basic HTTP GET request
	url := fmt.Sprintf("http://%v/meter/0", plug_host)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error getting response: ", err)
		return m, err
	}
	defer resp.Body.Close()

	// Read body from response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response: ", err)
		return m, err
	}
	// parse json response
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Println("ERROR parsing meter response: ", err)
		return m, err
	}

	// log.Println("   POWER: ", m.Power)
	log.Debug()
	return m, nil
}
