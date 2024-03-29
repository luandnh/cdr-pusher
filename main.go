package main

import (
	api "cdr-pusher/api"
	"cdr-pusher/internal/redis"
	"cdr-pusher/service"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Dir       string `env:"CONFIG_DIR" envDefault:"config/config.json"`
	Port      string
	LogType   string
	LogLevel  string
	LogFile   string
	DB        string
	Redis     string
	Notify    []string
	APICdrUrl string
}

var config Config

func init() {
	if err := env.Parse(&config); err != nil {
		log.Error("Get environment values fail")
		log.Fatal(err)
	}
	viper.SetConfigFile(config.Dir)
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err.Error())
		panic(err)
	}
	cfg := Config{
		Dir:       config.Dir,
		Port:      viper.GetString(`main.port`),
		LogType:   viper.GetString(`main.log_type`),
		LogLevel:  viper.GetString(`main.log_level`),
		LogFile:   viper.GetString(`main.log_file`),
		DB:        viper.GetString(`main.db`),
		Redis:     viper.GetString(`main.redis`),
		APICdrUrl: viper.GetString(`main.api_cdr_url`),
		Notify:    viper.GetStringSlice(`main.notify`),
	}
	if cfg.Redis == "enabled" {
		var err error
		redis.Redis, err = redis.NewRedis(redis.Config{
			Addr:         viper.GetString(`redis.address`),
			Password:     viper.GetString(`redis.password`),
			DB:           viper.GetInt(`redis.database`),
			PoolSize:     30,
			PoolTimeout:  20,
			IdleTimeout:  10,
			ReadTimeout:  20,
			WriteTimeout: 15,
		})
		if err != nil {
			panic(err)
		}
	}
	// for _, v := range cfg.Notify {
	// 	if v == "mail" {
	// 		service.SMTP_SERVER = viper.GetString("smtp.server")
	// 		service.SMTP_USERNAME = viper.GetString("smtp.username")
	// 		service.SMTP_PASSWORD = viper.GetString("smtp.password")
	// 		service.SMTP_RECEIVERS = viper.GetStringSlice("smtp.receivers")
	// 	}
	// }
	config = cfg
}

func main() {
	_ = os.Mkdir(filepath.Dir(config.LogFile), 0755)
	file, _ := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()
	setAppLogger(config, file)

	server := api.NewServer()
	//------ ADD CONTROLLER
	cdrService := service.NewCDR(config.APICdrUrl)
	api.APICdr(server.Engine, cdrService)

	s1 := gocron.NewScheduler(time.UTC)
	s1.SetMaxConcurrentJobs(1, gocron.RescheduleMode)
	s1.Every(30).Minutes().Do(cdrService.HandlePushBack)
	s1.StartAsync()
	defer s1.Clear()

	server.Start(config.Port)
}

func setAppLogger(cfg Config, file *os.File) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	// log.SetFormatter(&log.JSONFormatter{})
	switch cfg.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	switch cfg.LogType {
	case "DEFAULT":
		log.SetOutput(os.Stdout)
	case "FILE":
		if file != nil {
			log.SetOutput(io.MultiWriter(os.Stdout, file))
		} else {
			log.SetOutput(os.Stdout)
		}
	default:
		log.SetOutput(os.Stdout)
	}
}
