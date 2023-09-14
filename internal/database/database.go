package database

import (
	"fmt"
)

type Chrip struct {
	id int
	body string
}

type Database struct {
	chirps map[int]Chrip `json:"chirps"`
}

func SaveChirp(chirpBody string) error {
	fmt.Println("Saving chirp to database")
	return nil
}
