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
