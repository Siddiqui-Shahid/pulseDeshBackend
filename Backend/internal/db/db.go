package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"pulse-backend/internal/models"
)

var (
	mu     sync.RWMutex
	dbFile string
)

// Init initializes the database with a given file path.
func Init(path string) {
	mu.Lock()
	defer mu.Unlock()
	dbFile = path
}

// ReadUsers reads all users from the database file.
func ReadUsers() ([]models.User, error) {
	mu.RLock()
	defer mu.RUnlock()

	if dbFile == "" {
		return nil, fmt.Errorf("database not initialized")
	}

	content, err := ioutil.ReadFile(dbFile)
	if err != nil {
		// If file doesn't exist, return empty list
		if os.IsNotExist(err) {
			return []models.User{}, nil
		}
		return nil, err
	}

	var users []models.User
	if err := json.Unmarshal(content, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// WriteUsers writes all users to the database file.
func WriteUsers(users []models.User) error {
	mu.Lock()
	defer mu.Unlock()

	if dbFile == "" {
		return fmt.Errorf("database not initialized")
	}

	content, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dbFile, content, 0644)
}
