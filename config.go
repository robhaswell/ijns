package main

import (
	"encoding/json"
	"log"

	"github.com/deckarep/golang-set"
)

type CharacterConfig interface {
	// Return a set representing all the known characters that we are
	// interested in.
	CharacterSet() mapset.Set
	// Return the username to notify about a given character
	AlertUsername(string) string
}

type JsonCharacterConfig struct {
	// A set of all known characters
	characterSet mapset.Set
	// A map of character => username
	characterUsernameMap map[string]string
}

func NewJsonCharacterConfig(json string) (*JsonCharacterConfig, error) {
	config := &JsonCharacterConfig{}
	err := config.Init(json)
	return config, err
}

func (self *JsonCharacterConfig) Init(jsonString string) error {
	self.characterSet = mapset.NewSet()
	self.characterUsernameMap = make(map[string]string)

	characterMap := make(map[string][]string)

	if err := json.Unmarshal([]byte(jsonString), &characterMap); err != nil {
		return err
	}

	for username, characters := range characterMap {
		for _, character := range characters {
			self.characterSet.Add(character)
			self.characterUsernameMap[character] = username
		}
	}
	return nil
}

func (self *JsonCharacterConfig) CharacterSet() mapset.Set {
	return self.characterSet
}

func (self *JsonCharacterConfig) AlertUsername(character string) string {
	return self.characterUsernameMap[character]
}

// Initialise and return a JsonCharacterConfig of standardised test users,
func NewTestCharacterConfig() *JsonCharacterConfig {
	config, err := NewJsonCharacterConfig(`{"agrakari":["Maaya Saraki", "Indy Drone 4"], "fake_user":["Fake Character"]}`)
	if err != nil {
		log.Print(err)
	}
	return config
}
