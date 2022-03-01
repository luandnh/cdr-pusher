package main

import (
	/// THIRD PARTY PACKAGE

	"cdr-pusher/service"
	"time"

	IRedis "cdr-pusher/internal/redis"
	redis "cdr-pusher/internal/redis/driver"
	"io"
	"os"
	"path/filepath"

	/// PACKAGE OF POSTGRESQL

	/// PACKAGE OF RABBITMQ
	api "cdr-pusher/api"

	/// CALLBACK

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
	}
	if cfg.Redis == "enabled" {
		var err error
		IRedis.Redis, err = redis.NewRedis(redis.Config{
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
	config = cfg
}

func main() {
	_ = os.Mkdir(filepath.Dir(config.LogFile), 0755)
	file, _ := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()
	setAppLogger(config, file)

	server := api.NewServer()
	//------ ADD CONTROLLER
	cdrService := service.NewCdr(config.APICdrUrl)
	api.APICdr(server.Engine, cdrService)
	s4 := gocron.NewScheduler(time.UTC)
	s4.SetMaxConcurrentJobs(1, gocron.RescheduleMode)
	s4.Every(15).Seconds().Do(cdrService.HandlePushCdr)
	s4.StartAsync()
	defer s4.Clear()
	server.Start(config.Port)
}

func setAppLogger(cfg Config, file *os.File) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
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
			log.Error("main ", "Log File "+cfg.LogFile+" error")
			log.SetOutput(os.Stdout)
		}
	default:
		log.SetOutput(os.Stdout)
	}
}
