package main

import (
	"fmt"
	"github.com/configwizard/greenfinch-sdk/pkg/config"
)

func main() {
	config := config.ReadConfig()
	fmt.Printf("retrieved config %+v\r\n", config)
}
