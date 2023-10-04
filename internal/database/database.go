package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	Id int `json:"id"`
	Body string `json:"body"`
	AuthorId int `json:"author_id"`
}

type User struct {
	Id int `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
	IsChirpyRed bool `json:"is_chirpy_red"`
}

type Database struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
	RevokedTokens map[string]bool `json:"revokedTokens"`
}

const DbPath = "./database.json"
var dbLock = sync.Mutex{}

func loadDB() (Database, error) {
	var db Database

	raw, err := os.ReadFile(DbPath)
	if os.IsNotExist(err) {
		db = Database{ // TODO: don't like that we have to initialize this here, separately from definition
			Chirps: map[int]Chirp{},
			Users: map[int]User{},
			RevokedTokens: map[string]bool{},
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

func addChirp(chirpBody string, authorId int, db Database) (Chirp, Database) {
	id := len(db.Chirps) + 1
	chirp := Chirp{
		Id: id,
		Body: chirpBody,
		AuthorId: authorId,
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


func SaveChirp(chirpBody string, authorId int) (Chirp, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	db, err := loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, db := addChirp(chirpBody, authorId, db)
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

func UpdateUserMembership(id int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	user, ok := db.Users[id]
	if !ok {
		return errors.New("User not found")
	}

	user.IsChirpyRed = true
	db.Users[id] = user
	err = saveDB(db)
	return err
}

func UpdateUser(id int, email, password string) (User, error) {
	db, err := GetDB()
	if err != nil {
		return User{}, err
	}

	user, ok := db.Users[id]
	if !ok {
		return User{}, errors.New("User not found")
	}

	if email != "" {
		user.Email = email
	}
	if password != "" {
		const cost = 10 // TODO: need to consolidate password encryption
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		if err != nil {
			return User{}, err
		}
		user.Password = string(hashedPassword)
	}

	db.Users[id] = user
	err = saveDB(db)
	return user, err
}

func DeleteChirp(id int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, ok := db.Chirps[id]
	if !ok {
		return errors.New("Chirp not found")
	}

	delete(db.Chirps, id)
	err = saveDB(db)
	return err
}

func RevokedTokenExists(token string) (bool, error) {
	db, err := GetDB()
	if err != nil {
		return false, err
	}

	_, ok := db.RevokedTokens[token]
	return ok, nil
}

func AddRevokedToken(token string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	db.RevokedTokens[token] = true
	err = saveDB(db)
	return err
}

func GetDB() (Database, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	return loadDB()
}
