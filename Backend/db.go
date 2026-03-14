package main

import (
	"encoding/json"
	"os"
	"sync"
)

// dbMu guards all reads and writes to persons.json.
var dbMu sync.RWMutex

// readUsers loads the user list from the JSON file.
// Returns an empty slice when the file does not yet exist.
func readUsers() ([]User, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	data, err := os.ReadFile(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return []User{}, nil
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// writeUsers persists the user list to the JSON file (pretty-printed).
func writeUsers(users []User) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dbFile, data, 0o644)
}
