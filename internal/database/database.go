package database

import (
	"encoding/json"
	"os"
	"sync"
)

type Chirp struct {
	Id int `json:"id"`
	Body string `json:"body"`
}

type Database struct {
	Chirps map[int]Chirp `json:"chirps"`
}

const dbPath = "./database.json"
var dbLock = sync.Mutex{}

func loadDB() (Database, error) {
	var db Database

	raw, err := os.ReadFile(dbPath)
	if os.IsNotExist(err) {
		db = Database{
			Chirps: map[int]Chirp{},
		}
	} else {
		err = json.Unmarshal(raw, &db)

		if err != nil {
			return db, err
		}
	}

	return db, nil
}

func saveDB(db Database) error {
	raw, err := json.Marshal(db)
	if err != nil {
		return err
	}

	err = os.WriteFile(dbPath, raw, 0644)
	return err
}

func addChirp(chirpBody string, db Database) (Chirp, Database) {
	id := len(db.Chirps) + 1
	chirp := Chirp{
		Id: id,
		Body: chirpBody,
	}
	db.Chirps[id] = chirp
	return chirp, db
}


func SaveChirp(chirpBody string) (Chirp, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	db, err := loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, db := addChirp(chirpBody, db)
	err = saveDB(db)

	return chirp, err
}
