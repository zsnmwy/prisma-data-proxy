package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"wunderbase/pkg/graphiql"

	"github.com/buger/jsonparser"
	"github.com/wundergraph/graphql-go-tools/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/pkg/introspection"
	"go.uber.org/ratelimit"

	"github.com/go-redis/redis/v8"
)

var AdditionalConfig struct {
	ApiKey               string
	EnableRawQueries     bool
	EnableQueryEngineLog bool
	EnableMetrics bool
	QueryEngineHostBind string
	EnableOpenTelemetry bool
	OpenTelemetryEndpoint string
	EnableTelemetryInResponse bool
}

var RedisConfig struct {
	RedisEnable   bool
	RedisAddress  string
	RedisPassword string
	RedisDB       int
}

var rdb *redis.Client

type Handler struct {
	enableSleepMode   bool
	enablePlayground  bool
	queryEngineURL    string
	queryEngineSdlURL string
	healthEndpoint    string
	sleepAfterSeconds int
	init              sync.Once
	sleepCh           chan struct{}
	client            *http.Client
	readLimit         ratelimit.Limiter
	writeLimit        ratelimit.Limiter
	cancel            func()
}

func NewHandler(enableSleepMode bool, production bool, queryEngineURL string, queryEngineSdlURL, healthEndpoint string, sleepAfterSeconds, readLimitSeconds, writeLimitSeconds int, cancel func()) *Handler {
	if RedisConfig.RedisEnable {
		fmt.Println("Redis Enabled")
		rdb = redis.NewClient(&redis.Options{
			Addr:     RedisConfig.RedisAddress,
			Password: RedisConfig.RedisPassword,
			DB:       RedisConfig.RedisDB,
		})
	}

	return &Handler{
		enableSleepMode:   enableSleepMode,
		enablePlayground:  !production,
		queryEngineURL:    queryEngineURL,
		queryEngineSdlURL: queryEngineSdlURL,
		healthEndpoint:    healthEndpoint,
		sleepCh:           make(chan struct{}),
		sleepAfterSeconds: sleepAfterSeconds,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		readLimit:  ratelimit.New(readLimitSeconds),
		writeLimit: ratelimit.New(writeLimitSeconds),
		cancel:     cancel,
	}
}

type IntrospectionResponse struct {
	Data introspection.Data `json:"data"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	apiKeyFromQueryString := r.URL.Query().Get("api_key")
	if apiKeyFromQueryString == "" {
		apiKeyFromQueryString = r.URL.Query().Get("_token")
	}
	apiKeyFromHeader := r.Header.Get("authorization")

	if apiKeyFromQueryString != AdditionalConfig.ApiKey && apiKeyFromHeader != "Bearer "+AdditionalConfig.ApiKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.init.Do(func() {
		if h.enableSleepMode {
			go h.runSleepMode()
		}
		for {
			resp, err := http.Get(h.queryEngineURL)
			if err != nil || resp.StatusCode != http.StatusOK {
				time.Sleep(3 * time.Millisecond)
				continue
			}
			break
		}
	})

	if RedisConfig.RedisEnable && strings.HasPrefix(r.URL.Path, "/redis") {

		var arr []interface{}

		if r.Method == "POST" {
			err := json.NewDecoder(r.Body).Decode(&arr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if r.Method == "GET" {
			// /redis/set/key/value => ["SET", "key", "value"]
			parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/redis/"), "/")
			for _, part := range parts {
				arr = append(arr, part)
			}
		}

		w.Header().Add("Content-Type", "application/json")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result, err := rdb.Do(ctx, arr...).Result()

		var jsonResult []byte
		if err != nil {
			if err.Error() == "redis: nil" {
				jsonResult, _ = json.Marshal(map[string]interface{}{
					"result": nil,
				})
				w.WriteHeader(http.StatusOK)
			} else {
				jsonResult, _ = json.Marshal(map[string]interface{}{
					"error": err.Error(),
				})
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			jsonResult, _ = json.Marshal(map[string]interface{}{
				"result": result,
			})
			w.WriteHeader(http.StatusOK)
		}

		_, _ = w.Write(jsonResult)
		return
	}

	if r.URL.Path == h.healthEndpoint {
		// explicitly do this before the sleep mode check
		// otherwise the sleep mode will never be triggered
		resp, err := http.Get(h.queryEngineURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("query engine not reachable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		return
	}

	if h.enableSleepMode {
		defer func() {
			h.sleepCh <- struct{}{}
		}()
	}

	if h.enablePlayground && r.Header.Get("Content-Type") != "application/json" {
		w.Header().Add("Content-Type", "text/html")
		html := graphiql.GetGraphiqlPlaygroundHTML(r.RequestURI)
		_, _ = w.Write([]byte(html))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// check if body is introspection query
	if bytes.Contains(body, []byte("IntrospectionQuery")) {
		// if so, return the schema
		w.Header().Add("Content-Type", "application/json")
		gen := introspection.NewGenerator()
		// get the schema from the query engine on /sdl endpoint
		resp, err := http.Get(h.queryEngineSdlURL)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		schemaSDL, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		// generate the introspection result from the schema
		doc, report := astparser.ParseGraphqlDocumentBytes(schemaSDL)
		err = asttransform.MergeDefinitionWithBaseSchema(&doc)
		if err != nil {
			log.Fatalln(err)
		}
		var response IntrospectionResponse
		gen.Generate(&doc, &report, &response.Data)
		// marshal the result
		b, err := json.Marshal(response)
		if err != nil {
			log.Fatalln(err)
		}
		_, _ = w.Write(b)
		return
	}
	h.proxyRequestToEngine(body, w, r)
}

func (h *Handler) proxyRequestToEngine(body []byte, w http.ResponseWriter, r *http.Request) {
	variables, _, _, _ := jsonparser.Get(body, "variables")
	if variables == nil {
		// if no variables are set, set an empty object
		body, _ = jsonparser.Set(body, []byte("{}"), "variables")
	}
	operationName, _, _, _ := jsonparser.Get(body, "operationName")
	if operationName == nil {
		// if no operation name is set, set an empty string
		body, _ = jsonparser.Set(body, []byte("null"), "operationName")
	}
	for i := 0; i < 3; i++ {
		if h.sendRequest(body, w, r) {
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *Handler) sendRequest(body []byte, w http.ResponseWriter, r *http.Request) bool {

	if bytes.Contains(body, []byte("mutation")) {
		h.writeLimit.Take()
	}
	h.readLimit.Take()

	newRequest, err := http.NewRequestWithContext(r.Context(), r.Method, h.queryEngineURL, ioutil.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		log.Println(err)
		return false
	}
	// set the content type to application/json
	newRequest.Header.Set("content-type", "application/json")
	resp, err := h.client.Do(newRequest)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	if bytes.HasPrefix(data, []byte("{\"e")) && bytes.Contains(data, []byte("Timed out")) {
		return false
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (h *Handler) runSleepMode() {
	timer := time.NewTimer(time.Duration(h.sleepAfterSeconds) * time.Second)
	defer func() {
		fmt.Println("No requests for", h.sleepAfterSeconds, "seconds, cancelling context")
		h.cancel()
		return
	}()
	for {
		select {
		case <-h.sleepCh:
			done := timer.Reset(time.Duration(h.sleepAfterSeconds) * time.Second)
			if !done {
				return
			}
		case <-timer.C:
			return
		}
	}
}
