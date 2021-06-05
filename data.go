package main

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"

	bolt "go.etcd.io/bbolt"
)

const (
	PLUG_BUCKET = "PLUGS"
)

type PlugDescription struct {
	Id           string
	Hostname     string
	Name         string
	Type         string
	LastSeen     time.Time
	AddrV4       string
	Mac          string
	Is_available bool
}

func db_file_path() string {
	return viper.GetString("data.db_file")
}

func persist_record(measure Measure) {

	db, err := bolt.Open(db_file_path(), 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		// create bucket for that plug if needed
		b, err := tx.CreateBucketIfNotExists([]byte(measure.Id))
		if err != nil {
			return err
		}

		// now := time.Now()
		// nowFormat := now.Format(time.RFC3339)
		timeM := time.Unix(int64(measure.Timestamp), 0).Format(time.RFC3339)
		encoded, err := json.Marshal(measure)
		if err != nil {
			return err
		}

		err = b.Put([]byte(timeM), []byte(encoded))
		if err != nil {
			return fmt.Errorf("insert Measure: %s %s", timeM, err)
		}
		return err

	})
	if err != nil {
		log.Fatal("ERROR persist_record ", measure, err)
	}
}

// Save or update a plug information
func persist_plug(plug_desc PlugDescription) {

	db, err := bolt.Open(db_file_path(), 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		// create bucket for that plug if needed
		b, err := tx.CreateBucketIfNotExists([]byte(PLUG_BUCKET))
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(plug_desc)
		if err != nil {
			return err
		}

		err = b.Put([]byte(plug_desc.Mac), []byte(encoded))
		if err != nil {
			return fmt.Errorf("insert plug: %s %s", plug_desc.Hostname, err)
		}
		return err
	})
	if err != nil {
		log.Fatal("ERROR persist_plug ", plug_desc, err)
	}
}

// Get all known plugs
func get_plugs() (plugs []PlugDescription) {
	db, err := bolt.Open(db_file_path(), 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	plugs = make([]PlugDescription, 0, 10)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PLUG_BUCKET))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p PlugDescription
			err := json.Unmarshal(v, &p)
			if err != nil {
				return fmt.Errorf("Unmarshal json plug from db: %s %s", v, err)
			}
			plugs = append(plugs, p)
		}

		return nil
	})
	return plugs
}

func get_plug(plugId string) (plug PlugDescription) {

	db, err := bolt.Open(db_file_path(), 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PLUG_BUCKET))
		v := b.Get([]byte(plugId))
		err := json.Unmarshal(v, &plug)
		if err != nil {
			return fmt.Errorf("Unmarshal json plug from db: %s %s", v, err)
		}
		return nil
	})
	return
}

func updt_plug_availability(plugId string, is_available bool) {
	log.Debug("updt_plug_availability ", plugId)
	db, err := bolt.Open(db_file_path(), 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PLUG_BUCKET))
		v := b.Get([]byte(plugId))
		if v == nil {
			return nil
		}
		var plug PlugDescription
		errt := json.Unmarshal(v, &plug)
		if errt != nil {
			fmt.Println("ERROR", errt)
			// FIXME : trigerred when the plug is not found in db but ignored ?
			return fmt.Errorf("Unmarshal json plug from db: %s %s", v, errt)
		}

		plug.Is_available = is_available
		if is_available {
			plug.LastSeen = time.Now()
		}

		encoded, errt := json.Marshal(plug)
		if err != nil {
			return errt
		}

		errt = b.Put([]byte(plug.Mac), []byte(encoded))
		if err != nil {
			return fmt.Errorf("insert put plug: %s %s", plug.Hostname, err)
		}

		return nil

	})
	if err != nil {
		// FIXME : this exits the progem but no message is displayed ??
		log.Fatal("ERROR updating plug availability", plugId, err)
	}

}
