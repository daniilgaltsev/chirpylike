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

type User struct {
	Id int `json:"id"`
	Email string `json:"email"`
}

type Database struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
}

const dbPath = "./database.json"
var dbLock = sync.Mutex{}

func loadDB() (Database, error) {
	var db Database

	raw, err := os.ReadFile(dbPath)
	if os.IsNotExist(err) {
		db = Database{
			Chirps: map[int]Chirp{},
			Users: map[int]User{},
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

func addUser(email string, db Database) (User, Database) {
	id := len(db.Users) + 1
	user := User{
		Id: id,
		Email: email,
	}
	db.Users[id] = user
	return user, db
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

func SaveUser(email string) (User, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	db, err := loadDB()
	if err != nil {
		return User{}, err
	}

	user, db := addUser(email, db)
	err = saveDB(db)

	return user, err
}

func GetDB() (Database, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	return loadDB()
}
