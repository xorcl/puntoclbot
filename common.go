package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Poster interface {
	Start() error
	Post(string) error
}

type Monitor func(posters []Poster) error

func MustGetString(key string) (string, error) {
	s := viper.GetString(key)
	if len(s) == 0 {
		return "", fmt.Errorf("param with key %s was not found", key)
	}
	return s, nil
}
