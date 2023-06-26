# prisma-data-proxy

Self hosted [Prisma Data Proxy](https://www.prisma.io/docs/concepts/data-platform/data-proxy).

credits:

- https://github.com/wundergraph/graphql-go-tools
- https://github.com/wundergraph/wunderbase

other implementations:

- https://github.com/OnurGvnc/prisma-data-proxy-fastify
- https://github.com/aiji42/prisma-data-proxy-alt

## setup

- clone this repository
- update `schema.prisma`
- set environment variables. ([check out main.go](./main.go#L20-L39))
- `go run main.go`
- add the line `process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';` at the start of the node.js app
- detalils: [prisma data proxy](https://www.prisma.io/docs/concepts/data-platform/data-proxy).

## Metric

Access http://${QueryEnginePort}/metrics

## Env

| 变量名 | 类型 | 默认值 | 描述 |
| --- | --- | --- | --- |
| API_KEY | string | SECRET_API_KEY | Data Proxy Wrapper的API密钥 |
| PRODUCTION | bool | false | 是否在生产环境中运行 |
| ENABLE_SLEEP_MODE | bool | false | 是否启用睡眠模式 |
| SLEEP_AFTER_SECONDS | int | 10 | 进入睡眠模式前等待的秒数 |
| LISTEN_ADDR | string | 0.0.0.0:4466 | Data Proxy Wrapper监听的地址 |
| GRAPHIQL_API_URL | string | http://localhost:4466 | GraphiQL API的URL |
| READ_LIMIT_SECONDS | int | 10000 | 读取限制的秒数 |
| WRITE_LIMIT_SECONDS | int | 2000 | 写入限制的秒数 |
| HEALTH_ENDPOINT | string | /health | 健康检查的端点 |
| PRISMA_VERSION | string | 4bc8b6e1b66cb932731fb1bdbbc550d1e010de81 | Prisma的版本 |
| PRISMA_SCHEMA_FILE | string | ./schema.prisma | Prisma模式文件的路径 |
| ENABLE_MIGRATION | bool | false | 是否启用Prisma迁移 |
| MIGRATION_LOCK_FILE | string | migration.lock | 迁移锁文件的路径 |
| MIGRATION_ENGINE_PATH | string | ./migration-engine | 迁移引擎的路径 |
| QUERY_ENGINE_PATH | string | ./query-engine | 查询引擎的路径 |
| QUERY_ENGINE_PORT | string | 4467 | 查询引擎监听的端口 |
| QUERY_ENGINE_HOST_BIND | string | 127.0.0.1 | 查询引擎绑定的主机 |
| QUERY_ENGINE_LOG | bool | false | 是否记录查询引擎的日志 |
| QUERY_ENGINE_RAW_QUERIES | bool | true | 是否启用原始查询 |
| ENABLE_METRICS | bool | true | 是否启用Metric |
| ENABLE_OPEN_TELEMETRY | bool | false | 是否启用OpenTelemetry |
| OPEN_TELEMETRY_ENDPOINT | string |  | OpenTelemetry的Endpoint |
| ENABLE_TELEMETRY_IN_RESPONSE | bool | false | 是否在响应中启用遥测 |
| REDIS_REST_API_ENABLE | bool | false | 是否启用Redis REST API |
| REDIS_ADDRESS | string | localhost:6379 | Redis的地址 |
| REDIS_PASSWORD | string |  | Redis的密码 |
| REDIS_DB | int | 0 | Redis的数据库编号 |


## 🧪 Experimental Redis Rest API.

This project also serves as an experimental Redis REST API that is compatible with the [@upstash/redis](https://github.com/upstash/upstash-redis) package.

environment variables: ([check out main.go](./main.go#L41-L44))

```ts
import { Redis } from '@upstash/redis'

const redis = new Redis({
  url: 'http://localhost:4466/redis',
  token: 'SECRET_API_KEY',
  responseEncoding: false,
})
```
