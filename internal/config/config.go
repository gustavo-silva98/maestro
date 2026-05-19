package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	API struct {
		Name      string `mapstructure:"name"`
		Namespace string `mapstructure:"namespace"`
	} `mapstructure:"api"`
	Jira struct {
		Name          string `mapstructure:"name"`
		Token         string `mapstructure:"token"`
		UserName      string `mapstructure:"username"`
		TenantName    string `mapstructure:"tenant_name"`
		StatusAllowed struct {
			InitialStatus string `mapstructure:"initial_status"`
			FinalStatus   string `mapstructure:"final_status"`
		} `mapstructure:"status_allowed"`
	} `mapstructure:"jira"`
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.SetEnvPrefix("MAESTRO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("erro lendo config: %w", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("erro parseando config: %w", err)
	}
	return cfg, nil
}
