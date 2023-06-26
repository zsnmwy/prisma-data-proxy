#!/bin/sh

echo DATABASE_URL $DATABASE_URL
echo DATA_PROXY_API_KEY $DATA_PROXY_API_KEY
yarn install
if [ -n "$MIGRATE" ]; then
  yarn prisma migrate reset -f
fi
exec "$@"
