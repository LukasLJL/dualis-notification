package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Username              string `mapstructure:"DUALIS_USER"`
	Password              string `mapstructure:"DUALIS_PASSWORD"`
	UpdateIntervalMinutes int    `mapstructure:"INTERVAL"`
}

type SMTPConfig struct {
	SMTPHost              string `mapstructure:"SMTP_HOST"`
	SMTPPort              int    `mapstructure:"SMTP_PORT"`
	SMTPUsername          string `mapstructure:"SMTP_USER"`
	SMTPPassword          string `mapstructure:"SMTP_PASSWORD"`
	NotificationRecipient string `mapstructure:"SMTP_RECEIVER"`
}

var Dualis Config
var SMTP SMTPConfig

func GetConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("Error reading config file")
		panic(err)
	}

	err_dualis := viper.Unmarshal(&Dualis)
	err_smtp := viper.Unmarshal(&SMTP)

	if err_dualis != nil || err_smtp != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}
