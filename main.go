package main

import (
	"fmt"
	"log"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

func main() {
	log.Printf("Starting bot...")
	posters := []Poster{
		&TelegramBot{},
	}
	for _, poster := range posters {
		err := poster.Start()
		if err != nil {
			panic(err)
		}
	}
	c := cron.New()
	c.AddFunc("* * * * *", func() {
		log.Printf("Executing new domains task...")
		err := monitorNewDomains(posters)
		if err != nil {
			log.Printf("Error executing new domains: %s", err)
		}
	})
	c.AddFunc("* * * * *", func() {
		log.Printf("Executing arbitraje task...")
		err := monitorArbitraje(posters)
		if err != nil {
			log.Printf("Error executing new domains: %s", err)
		}
	})
	c.AddFunc("41 0 * * *", func() {
		log.Printf("Executing deleted domains task...")
		err := monitorDeletedDomains(posters)
		if err != nil {
			log.Printf("Error executing deleted domains: %s", err)
		}
	})
	c.Run()
}
