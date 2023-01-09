package main

import (
	"github.com/lukasljl/dualis-notification/config"
	"github.com/lukasljl/dualis-notification/dualis"
)

func main() {
	config.GetConfig()
	dualis.InitDualis()
}
