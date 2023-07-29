FROM golang:1.19.0-alpine as builder

ENV PRISMA_VERSION="6b0aef69b7cdfc787f822ecd7cdc76d5f1991584"
ENV OS="linux-musl"
ENV QUERY_ENGINE_URL="https://binaries.prisma.sh/all_commits/${PRISMA_VERSION}/${OS}/query-engine.gz"

# install prisma
WORKDIR /app/prisma
# download query engine
RUN wget -O query-engine.gz $QUERY_ENGINE_URL
RUN gunzip query-engine.gz
RUN chmod +x query-engine

# build app
WORKDIR /app
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main .

FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/prisma/query-engine /app/query-engine
COPY ./schema.prisma .
RUN chmod +x /app/query-engine
RUN apk add openssl1.1-compat postgresql-client bash openssl libgcc libstdc++ ncurses-libs lsof curl --no-cache
ENV QUERY_ENGINE_PATH="/app/query-engine"
ENV PRISMA_SCHEMA_FILE="/app/schema.prisma"
RUN mkdir /app/data
EXPOSE 4466
ENTRYPOINT ["/app/main"]