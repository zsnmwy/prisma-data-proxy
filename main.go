package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"wunderbase/pkg/api"
	"wunderbase/pkg/migrate"
	"wunderbase/pkg/queryengine"

	"github.com/caarlos0/env/v6"
)

type config struct {
	// Data Proxy Wrapper
	ApiKey            string `env:"API_KEY" envDefault:"SECRET_API_KEY"`
	Production        bool   `env:"PRODUCTION" envDefault:"false"`
	EnableSleepMode   bool   `env:"ENABLE_SLEEP_MODE" envDefault:"false"`
	SleepAfterSeconds int    `env:"SLEEP_AFTER_SECONDS" envDefault:"10"`
	ListenAddr        string `env:"LISTEN_ADDR" envDefault:"0.0.0.0:4466"`
	GraphiQLApiURL    string `env:"GRAPHIQL_API_URL" envDefault:"http://localhost:4466"`
	ReadLimitSeconds  int    `env:"READ_LIMIT_SECONDS" envDefault:"10000"`
	WriteLimitSeconds int    `env:"WRITE_LIMIT_SECONDS" envDefault:"2000"`
	HealthEndpoint    string `env:"HEALTH_ENDPOINT" envDefault:"/health"`

	// Prisma
	PrismaVersion string `env:"PRISMA_VERSION" envDefault:"4bc8b6e1b66cb932731fb1bdbbc550d1e010de81"` // 4.4.0-29 @ https://github.com/prisma/engines-wrapper/blob/main/packages/engines-version/package.json

	// Prisma Migration Engine - Schema
	PrismaSchemaFilePath  string `env:"PRISMA_SCHEMA_FILE" envDefault:"./schema.prisma"`
	EnableMigration       bool   `env:"ENABLE_MIGRATION" envDefault:"false"`
	MigrationLockFilePath string `env:"MIGRATION_LOCK_FILE" envDefault:"migration.lock"`
	MigrationEnginePath   string `env:"MIGRATION_ENGINE_PATH" envDefault:"./migration-engine"`

	// I think that we should discard `EnablePlayground`, when we add `Production` flag.
	// EnablePlayground      bool   `env:"ENABLE_PLAYGROUND" envDefault:"true"`

	// Prisma Query Engine - Instance
	QueryEnginePath     string `env:"QUERY_ENGINE_PATH" envDefault:"./query-engine"`
	QueryEnginePort     string `env:"QUERY_ENGINE_PORT" envDefault:"4467"`
	QueryEngineHostBind string `env:"QUERY_ENGINE_HOST_BIND" envDefault:"127.0.0.1"`
	QueryEngineLog      bool   `env:"QUERY_ENGINE_LOG" envDefault:"false"`
	EnableRawQueries    bool   `env:"QUERY_ENGINE_RAW_QUERIES" envDefault:"true"`
	// Prisma Query Engine - Trace && Metrics
	EnableMetrics         bool `env:"ENABLE_METRICS" envDefault:"true"`
	EnableOpenTelemetry   bool `env:"ENABLE_OPEN_TELEMETRY" envDefault:"false"`
	OpenTelemetryEndpoint string `env:"OPEN_TELEMETRY_ENDPOINT" envDefault:""`
	EnableTelemetryInResponse bool `env:"ENABLE_TELEMETRY_IN_RESPONSE" envDefault:"false"`

	// Redis Config
	RedisRestAPIEnable bool   `env:"REDIS_REST_API_ENABLE" envDefault:"false"`
	RedisAddress       string `env:"REDIS_ADDRESS" envDefault:"localhost:6379"`
	RedisPassword      string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB            int    `env:"REDIS_DB" envDefault:"0"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	config := &config{}
	if err := env.Parse(config); err != nil {
		log.Fatalln("parse env", err)
	}
	api.AdditionalConfig.ApiKey = config.ApiKey
	api.AdditionalConfig.EnableRawQueries = config.EnableRawQueries
	api.AdditionalConfig.EnableQueryEngineLog = config.QueryEngineLog
	api.AdditionalConfig.EnableMetrics = config.EnableMetrics
	api.AdditionalConfig.QueryEngineHostBind = config.QueryEngineHostBind
	api.AdditionalConfig.EnableTelemetryInResponse = config.EnableTelemetryInResponse
	api.AdditionalConfig.OpenTelemetryEndpoint = config.OpenTelemetryEndpoint
	api.AdditionalConfig.EnableOpenTelemetry = config.EnableOpenTelemetry
	api.RedisConfig.RedisEnable = config.RedisRestAPIEnable
	api.RedisConfig.RedisAddress = config.RedisAddress
	api.RedisConfig.RedisPassword = config.RedisPassword
	api.RedisConfig.RedisDB = config.RedisDB
	schema, err := ioutil.ReadFile(config.PrismaSchemaFilePath)
	if err != nil {
		log.Fatalln("load prisma schema", err)
	}
	if config.EnableMigration {
		migrate.Database(config.MigrationEnginePath, config.MigrationLockFilePath, string(schema), config.PrismaSchemaFilePath)
	}
	go queryengine.Run(ctx, wg, config.QueryEnginePath, config.QueryEnginePort, config.PrismaSchemaFilePath, config.Production)
	log.Printf("Server Listening on: http://%s", config.ListenAddr)
	handler := api.NewHandler(config.EnableSleepMode,
		config.Production,
		fmt.Sprintf("http://localhost:%s/", config.QueryEnginePort),
		fmt.Sprintf("http://localhost:%s/sdl", config.QueryEnginePort),
		config.HealthEndpoint,
		config.SleepAfterSeconds,
		config.ReadLimitSeconds,
		config.WriteLimitSeconds,
		cancel)
	srv := http.Server{
		Addr:    config.ListenAddr,
		Handler: handler,
	}
	go func() {
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalln("listen and serve", err)
		}
	}()
	<-ctx.Done()
	err = srv.Close()
	if err != nil {
		log.Fatalln("close server", err)
	}
	log.Println("Server stopped")
	wg.Done()
	wg.Wait()
	os.Exit(0)
}
