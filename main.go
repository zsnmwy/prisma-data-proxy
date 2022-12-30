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
	Production            bool   `env:"PRODUCTION" envDefault:"false"`
	PrismaSchemaFilePath  string `env:"PRISMA_SCHEMA_FILE" envDefault:"./schema.prisma"`
	MigrationLockFilePath string `env:"MIGRATION_LOCK_FILE" envDefault:"migration.lock"`
	EnableSleepMode       bool   `env:"ENABLE_SLEEP_MODE" envDefault:"true"`
	SleepAfterSeconds     int    `env:"SLEEP_AFTER_SECONDS" envDefault:"10"`
	// I think that we should discard `EnablePlayground`, when we add `Production` flag.
	// EnablePlayground      bool   `env:"ENABLE_PLAYGROUND" envDefault:"true"`
	MigrationEnginePath string `env:"MIGRATION_ENGINE_PATH" envDefault:"./migration-engine"`
	QueryEnginePath     string `env:"QUERY_ENGINE_PATH" envDefault:"./query-engine"`
	QueryEnginePort     string `env:"QUERY_ENGINE_PORT" envDefault:"4467"`
	ListenAddr          string `env:"LISTEN_ADDR" envDefault:"0.0.0.0:4466"`
	GraphiQLApiURL      string `env:"GRAPHIQL_API_URL" envDefault:"http://localhost:4466"`
	ReadLimitSeconds    int    `env:"READ_LIMIT_SECONDS" envDefault:"10000"`
	WriteLimitSeconds   int    `env:"WRITE_LIMIT_SECONDS" envDefault:"2000"`
	HealthEndpoint      string `env:"HEALTH_ENDPOINT" envDefault:"/health"`
	EnableMigration     bool   `env:"ENABLE_MIGRATION" envDefault:"false"`
	QueryEngineLog      bool   `env:"QUERY_ENGINE_LOG" envDefault:"false"`
	EnableRawQueries    bool   `env:"QUERY_ENGINE_RAW_QUERIES" envDefault:"true"`
	PrismaVersion       string `env:"PRISMA_VERSION" envDefault:"fb56bfedb3a66723ec5d908bc5cb4a411ec97c47"` // 4.4.0-29 @ https://github.com/prisma/engines-wrapper/blob/main/packages/engines-version/package.json
	ApiKey              string `env:"API_KEY" envDefault:"SECRET_API_KEY"`

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
