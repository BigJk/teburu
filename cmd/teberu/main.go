package main

import (
	"context"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
	"os/signal"
	"teburu"
	"time"
)

const (
	ConfigCredentialsFile = "credentials_file"
	ConfigCors            = "cors"
	ConfigBind            = "bind"
	ConfigRateLimit       = "rate_limit"
	ConfigCaching         = "cache"
	ConfigCacheTTL        = "cache_ttl"
)

func setupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault(ConfigCredentialsFile, "./creds.json")
	viper.SetDefault(ConfigCors, true)
	viper.SetDefault(ConfigBind, ":8753")
	viper.SetDefault(ConfigRateLimit, 5.0)
	viper.SetDefault(ConfigCaching, false)
	viper.SetDefault(ConfigCacheTTL, 5*time.Minute)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_ = viper.SafeWriteConfigAs("./config.yaml")
		} else {
			panic(err)
		}
	}

	log.Info("Loaded config", "file", viper.ConfigFileUsed())
}

func main() {
	setupConfig()

	// Setup Google Sheets API
	ctx := context.Background()
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(viper.GetString(ConfigCredentialsFile)))
	if err != nil {
		log.Fatal("Failed to create service", "error", err)
	}

	// Setup server
	server := teburu.NewServer(service)

	if viper.GetBool(ConfigCors) {
		server.EnableCORS()
	}

	if viper.GetInt(ConfigRateLimit) > 0 {
		server.SetupRateLimit(viper.GetFloat64(ConfigRateLimit))
	}

	if viper.GetBool(ConfigCaching) {
		server.EnableCaching(viper.GetDuration(ConfigCacheTTL))
	}

	// Start server
	go func() {
		if err := server.Start(viper.GetString(ConfigBind)); err != nil && err != http.ErrServerClosed {
			log.Fatal("shutting down the server", "error", err)
		}
	}()

	// Ctrl+C Handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
