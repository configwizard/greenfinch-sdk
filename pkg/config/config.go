package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Peers map[string]Peer `yaml:"peers"`
}

type Peer struct {
	Address string `yaml:"address"`
	Priority int `yaml:"priority"`
	Weight int `yaml:"weight"`
}

func ReadConfig() Config {
	viper.SetConfigName("cfg") // name of config file (without extension)
	viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	//viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.AddConfigPath("/Users/alex.walker/go/src/github.com/configwizard/greenfinch-sdk/")               // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Fatal("could not find the config file specified ", err)
		} else {
			// Config file was found but another error was produced
			log.Fatal("error reading config file ", err)
		}
	}
	conf := &Config{}
	if err := viper.Unmarshal(conf); err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}
	for k, v := range conf.Peers {
		fmt.Println(k, v)
	}

	return *conf
}