# prisma-data-proxy

Self hosted [Prisma Data Proxy](https://www.prisma.io/docs/concepts/data-platform/data-proxy).

credits:

- https://github.com/wundergraph/graphql-go-tools
- https://github.com/wundergraph/wunderbase

other implementations:

- https://github.com/OnurGvnc/prisma-data-proxy-fastify
- https://github.com/aiji42/prisma-data-proxy-alt

## TLDR

Most of the time you don't need to care about this repo.

Please to see [the guide repo](https://github.com/zsnmwy/prisma-data-proxy-windmill-template).

It can works with Prisma `4.x.x-5.x.x`. **But you must make the query engine to match Prisma version**.

The guide repo is already resolve this question.

It will check Prisma Version and download the correct version when build the docker image.

**Please don't use the Migration Engine in this repo.**

**Best practics ref the guide repo.**

## About Prisma Query Engine Version

```bash
yarn prisma version
```

```log
root@6b4ebdef442f:/app# yarn prisma version
yarn run v1.22.19
$ /app/node_modules/.bin/prisma version
prisma                  : 5.0.0
@prisma/client          : 5.0.0
Current platform        : debian-openssl-1.1.x
Query Engine (Node-API) : libquery-engine 6b0aef69b7cdfc787f822ecd7cdc76d5f1991584 (at node_modules/@prisma/engines/libquery_engine-debian-openssl-1.1.x.so.node)
Schema Engine           : schema-engine-cli 6b0aef69b7cdfc787f822ecd7cdc76d5f1991584 (at node_modules/@prisma/engines/schema-engine-debian-openssl-1.1.x)
Schema Wasm             : @prisma/prisma-schema-wasm 4.17.0-26.6b0aef69b7cdfc787f822ecd7cdc76d5f1991584
Default Engines Hash    : 6b0aef69b7cdfc787f822ecd7cdc76d5f1991584
Studio                  : 0.487.0
Done in 1.30s.
```

The Query Engine version is `6b0aef69b7cdfc787f822ecd7cdc76d5f1991584`.

## setup (Deprecated)

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


## Prisma 5.0 jsonProtocol

If you want to use data proxy with `Prisma 5.0`, just only upgrade the query engine to match the version `yarn prisma version`.

### Client Request To Query Engine

> Code src/index.tsx

POST /5.0.0/0c851fb4212291ec29bc4fbc42b1fbb7a42875e19c540ee0e7aae40496d7e23e/graphql HTTP/1.0

```json
{
  "modelName": "User",
  "action": "createOne",
  "query": {
    "arguments": {
      "data": {
        "email": "2023-07-29T05:51:33.569Z123@email.com",
        "posts": {
          "create": {
            "title": "posts",
            "attr": {
              "a": 123,
              "b": true,
              "c": {
                "d": 22,
                "e": "123"
              }
            }
          }
        }
      }
    },
    "selection": {
      "$composites": true,
      "$scalars": true
    }
  }
}
```

### Response
```json
{
  "data": {
    "createOneUser": {
      "id": 1,
      "email": "2023-07-29T05:51:33.569Z123@email.com",
      "name": null
    }
  }
}
```

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
