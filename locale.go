package main

import (
	"encoding/json"
	"io/ioutil"
)

type Locale struct {
	Translations map[string]string
}

func LoadLocale(path string) (*Locale, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var translations map[string]string
	err = json.Unmarshal(file, &translations)
	return &Locale{translations}, err
}
