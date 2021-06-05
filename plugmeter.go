package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
)

const (
	PLUG_POLL_PERIOD = 2
)

func main() {
	init_configuration()
	configure_log()
	print_configuration()

	plug_events := make(chan PlugEvent)

	if viper.GetBool("plugs.discovery") {
		go continuous_plug_detection(10, plug_events)
	}

	measurements := make(chan Measure, 20)
	go plug_monitor(plug_events, measurements)

	// manually inject PLUG_ARRIVAL events for plug with static conf
	for index, ip := range viper.GetStringSlice("plugs.ips") {
		log.Debug("Injecting static IP ", ip, index)

		plug := PlugEntry{
			DetectionId: fmt.Sprintf("static_id_%d", index),
			Id:          fmt.Sprintf("static_id_%d", index),
			AddrV4:      net.ParseIP(ip),
			AddrV6:      nil,
		}

		plug_events <- PlugEvent{
			EventType: PLUG_ARRIVAL,
			Plug:      plug,
		}
	}

	go store_measurements(measurements)

	go start_webui(viper.GetInt("web_ui.port"))

	// This will keep the goroutines running until
	// the program is stopped
	select {}

}

// Setup configuration mechanism and default values.
// Configuration can be set with a toml file or CLI flags.
func init_configuration() {

	viper.SetDefault("logs.level", "debug")
	viper.SetDefault("web_ui.port", 3000)
	viper.SetDefault("plugs.discovery", true)
	viper.SetDefault("plugs.ips", []string{})
	viper.SetDefault("plugs.poll_period", 2)
	viper.SetDefault("plugs.max_error", 2)
	viper.SetDefault("data.csv", true)
	viper.SetDefault("data.csv_file", "plugmeter.csv")
	viper.SetDefault("data.discovery", "plugmeter.db")

	// CLI flag configuration
	var ui_port int
	flag.IntVar(&ui_port, "port", 3000, "port fro Web UI (default: 3000)")
	var plug_discovery bool
	flag.BoolVar(&plug_discovery, "discovery", true, "use mDNS to discover plugs")
	var plug_ips []string
	flag.StringArrayVar(&plug_ips, "plug_ip", nil, "Plugs static IPs")
	var log_level string
	flag.StringVar(&log_level, "log", "warning", "Log level")
	var period int
	flag.IntVar(&period, "period", 3000, "Number of second between two measurements on each plug")
	var max_error int
	flag.IntVar(&max_error, "max_error", 3000, "Number of errors before considering a plug to be unavailable")
	var out_csv bool
	flag.BoolVar(&out_csv, "csv", false, "Output energy measurements to a csv file")
	var out_csv_file string
	flag.StringVar(&out_csv_file, "csv_file", "plugmeter.csv", "Output csv file")
	var out_db_file string
	flag.StringVar(&out_db_file, "db_file", "plugmeter.db", "Output DB file")
	var conf_path string
	flag.StringVar(&conf_path, "conf", "none", "configuration file path")

	flag.Parse()

	// Config file configuration
	if conf_path != "none" {
		var c_dir, c_file = filepath.Split(conf_path)
		viper.SetConfigFile(c_file)
		viper.AddConfigPath(c_dir)
	} else {
		viper.SetConfigName("plugmeter_conf")
		viper.SetConfigType("toml")
		viper.AddConfigPath("/etc/plugmeter/")
		viper.AddConfigPath("$HOME/.plugmeter")
		viper.AddConfigPath(".")
	}
	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Error(err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("No config file ")
		} else {
			log.Fatalf("Fatal error reading config file: %s ", err)
		}
	} else {
		log.Info("Config file found")
	}

	viper.BindPFlag("web_ui.port", flag.Lookup("port"))
	viper.BindPFlag("plugs.discovery", flag.Lookup("discovery"))
	viper.BindPFlag("plugs.ips", flag.Lookup("plug_ip"))
	viper.BindPFlag("plugs.poll_period", flag.Lookup("period"))
	viper.BindPFlag("plugs.max_error", flag.Lookup("max_error"))
	viper.BindPFlag("logs.level", flag.Lookup("log"))
	viper.BindPFlag("data.csv", flag.Lookup("csv"))
	viper.BindPFlag("data.csv_file", flag.Lookup("csv_file"))
	viper.BindPFlag("data.db_file", flag.Lookup("db_file"))

	// configuration with ENV variables
	viper.BindEnv("logs.level", "LOG_LEVEL")
	viper.BindEnv("web_ui.port", "UI_PORT")
	viper.BindEnv("plugs.discovery", "PLUG_DISCOVERY")
	viper.BindEnv("plugs.ips", "PLUG_IPS")
	viper.BindEnv("plugs.poll_period", "POLL_PERIOD")
	viper.BindEnv("plugs.max_error", "MAX_ERROR")
	viper.BindEnv("data.csv", "CSV_OUT")
	viper.BindEnv("data.csv_file", "CSV_FILE")
	viper.BindEnv("data.db_file", "DB_FILE")
}

func print_configuration() {
	log.Debug("**** Using configuration ****")
	log.Debug("*  Log_levels: ", viper.Get("logs.level"))
	log.Debug("*  Web UI port: ", viper.Get("web_ui.port"))
	log.Debug("*  Plug Detection: ", viper.Get("plugs.discovery"))
	log.Debug("*  Plug IPs: ", viper.Get("plugs.ips"))
	log.Debug("*  Poll period: ", viper.Get("plugs.poll_period"))
	log.Debug("*  Max error: ", viper.Get("plugs.max_error"))
	log.Debug("*  CSV output: ", viper.Get("data.csv"))
	log.Debug("*  CSV output file: ", viper.Get("data.csv_file"))
	log.Debug("*  DB file: ", viper.Get("data.db_file"))
	log.Debug("*****************************")
}

func configure_log() {
	switch viper.Get("logs.level") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Warn("Invalid log level %s, using 'info' instead", viper.GetString("logs.level"))
		log.SetLevel(log.InfoLevel)
	}
}

// Start and stops monitoring plug dependeing on detection events.
func plug_monitor(plug_events chan PlugEvent, measurements chan Measure) {
	done := make(chan bool)
	plugs := make(map[string]bool)
	for {
		select {
		case e := <-plug_events:
			if e.EventType == PLUG_ARRIVAL {
				if !plugs[e.Plug.DetectionId] {
					log.Info("Starting polling for: ", e.Plug.Id, e.Plug.AddrV4)
					go poll_plug(e.Plug, done, measurements, plug_events)
					plugs[e.Plug.DetectionId] = true
				}
			} else if e.EventType == PLUG_REMOVAL {
				log.Info("REMOVE %s from available plugs", e.Plug.Id)
				// No need to stop polling: automatic
				// simply remove from current list of plug,
				// to be able to restart polling later
				delete(plugs, e.Plug.DetectionId)
				updt_plug_availability(e.Plug.Id, false)
			}
		}
	}
}

// Listen for measurement on the `measurements` channel
// and store them.
func store_measurements(measurements chan Measure) {
	// Read and store the measurements
	for m := range measurements {
		// fmt.Println("Measure : ", m)
		persist_record(m)
		log_measurements_csv(m)
	}
}

func log_measurements_csv(m Measure) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(viper.GetString("data.csv_file"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open csv file file for writting '%s' ", viper.GetString("data.csv_file"), err)
	}
	defer f.Close()

	timeM := time.Unix(int64(m.Timestamp), 0).Format(time.RFC3339)
	record := []string{m.Id, m.Plug, strconv.FormatUint(m.Timestamp, 10), timeM, strconv.FormatFloat(m.Power, 'f', 6, 64), strconv.FormatInt(int64(m.Energy), 10)}
	writer := csv.NewWriter(f)
	writer.Write(record)
	writer.Flush()
}

type Measure struct {
	Id        string
	Power     float64
	Energy    uint32
	Plug      string
	Timestamp uint64
}

// Polls a plug periodically to get energy consumption.
// First get a full description of the plug
// and then start polling every `PLUG_POLL_PERIOD` seconds,
// sending measurements on the `measurement` channel.
// If a plug cannot be reached `MAX_ERROR_COUNT` times consecutively,
// it is considered as removed and a corresponding `PlugEvent` is sent on
// the `plug_events` channel.
func poll_plug(plugDetection PlugEntry, done chan bool,
	measurements chan Measure, plug_events chan PlugEvent) {

	plug_desc, err := get_plug_desc(plugDetection.AddrV4.String())
	if err == nil {
		log.Debugf("Initial plug info: %s", plug_desc)
	} else {
		log.Warnf("Could not get plug info at %s ", plugDetection, err)
		plug_events <- PlugEvent{
			EventType: PLUG_REMOVAL,
			Plug: PlugEntry{
				DetectionId: plugDetection.Id,
				AddrV4:      plugDetection.AddrV4,
			},
		}
		return
	}
	persist_plug(plug_desc)

	ticker := time.NewTicker(time.Duration(viper.GetInt("plugs.poll_period")) * time.Second)
	defer ticker.Stop()

	error_count := 0
	for {
		select {
		case <-done:
			log.Info("Stopping polling plugs")
			return
		case t := <-ticker.C:
			m, err := get_energy_data(plugDetection.AddrV4.String())
			if err != nil {
				log.Infof("- %s COULD not get power at %v %v %s\n", plugDetection, t, error_count, err)
				error_count++
			} else {
				// fmt.Printf("- %s %f %s \n ", plug_host, m.Power, t)
				var measure = Measure{
					Id:        plug_desc.Mac,
					Power:     m.Power,
					Energy:    m.Total,
					Plug:      plugDetection.AddrV4.String(),
					Timestamp: m.Timestamp}
				measurements <- measure

				updt_plug_availability(plug_desc.Id, true)
				error_count = 0
			}
		}
		if error_count > viper.GetInt("plugs.max_error") {
			log.Warnf("Could not reach %s, stopping polling", plugDetection)
			plug_events <- PlugEvent{
				EventType: PLUG_REMOVAL,
				Plug: PlugEntry{
					Id:          plug_desc.Id,
					DetectionId: plugDetection.Id,
					AddrV4:      plugDetection.AddrV4,
				},
			}
			break
		}
	}
}

// enum-like type for plug detection events:
type PlugEventType uint

const (
	PLUG_ARRIVAL PlugEventType = iota
	PLUG_REMOVAL
)

type PlugEvent struct {
	EventType PlugEventType
	Plug      PlugEntry
}

// Run mDns plug detection periodically every `period` seconds
// and emit a PlugEvent on the `plug_detection` channel for
// each detected plug.
func continuous_plug_detection(period int, plug_detection chan PlugEvent) {
	log.Debug("Discovering new plugs")

	ticker := time.NewTicker(time.Duration(period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			plugs := detectPlugs()
			for _, p := range plugs {
				plug_detection <- PlugEvent{
					EventType: PLUG_ARRIVAL,
					Plug:      p,
				}
			}
		}
	}

}
