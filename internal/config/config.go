package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Mode string `mapstructure:"mode"`
	API  struct {
		Name           string `mapstructure:"name"`
		Namespace      string `mapstructure:"namespace"`
		Port           string `mapstructure:"port"`
		ConcurrentJobs int    `mapstructure:"concurrentJobs"`
	} `mapstructure:"api"`
	Jira struct {
		Name          string `mapstructure:"name"`
		Token         string `mapstructure:"token"`
		UserName      string `mapstructure:"username"`
		TenantName    string `mapstructure:"tenant_name"`
		WebhookSecret string `mapstructure:"webhook_secret"`
		StatusAllowed struct {
			InitialStatus string `mapstructure:"initial_status"`
			FinalStatus   string `mapstructure:"final_status"`
		} `mapstructure:"status_allowed"`
	} `mapstructure:"jira"`
	DB struct {
		Name     string `mapstructure:"name"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DB       string `mapstructure:"db"`
		URL      string `mapstructure:"url"`
	} `mapstructure:"database"`
}

type JobType struct {
	JobType    string `mapstructure:"jobType"`
	JiraFields struct {
		Need                  string `mapstructure:"need"`
		System                string `mapstructure:"system"`
		AttachmentFilename    string `mapstructure:"attachmentFilename"`
		AnswerCommentTemplate string `mapstructure:"answerCommentTemplate"`
	} `mapstructure:"jiraFields"`
	Container struct {
		ImageName    string `mapstructure:"imageName"`
		ContainerDir string `mapstructure:"containerDir"`
		Cpus         int    `mapstructure:"cpus"`
		Memory       string `mapstructure:"memory"`
	} `mapstructure:"container"`
	Indicators struct {
		TimerPerOp int `mapstructure:"timePerOp"`
		TimeSaved  int `mapstructure:"timesaved"`
	}
}

func LoadConfig() (Config, error) {

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
	if cfg.Mode == "dev" {
		return cfg, nil
	}
	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("erro lendo .env: %w", err)
	}
	return cfg, nil
}

func LoadJobTypes(dir string) (map[string]JobType, error) {
	f, err := os.ReadDir(dir)
	if err != nil {
		return map[string]JobType{}, err
	}
	types := make(map[string]JobType)
	for _, file := range f {
		v := viper.New()
		v.SetConfigFile(filepath.Join(dir, file.Name()))
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("Falha ao ler configurações: %v", err)
		}
		var jt JobType
		if err := v.Unmarshal(&jt); err != nil {
			return nil, fmt.Errorf("erro ao decodificar %s: %w", file.Name(), err)
		}
		types[jt.JobType] = jt
	}
	return types, nil
}
