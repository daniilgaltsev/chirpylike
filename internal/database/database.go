package database

import (
	"encoding/json"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	Id int `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id int `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type Database struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
}

const DbPath = "./database.json"
var dbLock = sync.Mutex{}

func loadDB() (Database, error) {
	var db Database

	raw, err := os.ReadFile(DbPath)
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

	err = os.WriteFile(DbPath, raw, 0644)
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

func addUser(email, password string, db Database) (User, Database) {
	id := len(db.Users) + 1
	user := User{
		Id: id,
		Email: email,
		Password: password,
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

func SaveUser(email, password string) (User, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	db, err := loadDB()
	if err != nil {
		return User{}, err
	}

	const cost = 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return User{}, err
	}
	user, db := addUser(email, string(hashedPassword), db)
	err = saveDB(db)

	return user, err
}

func GetDB() (Database, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	return loadDB()
}
