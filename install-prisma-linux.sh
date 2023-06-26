# define variable
PRISMA_VERSION="4bc8b6e1b66cb932731fb1bdbbc550d1e010de81"
OS="linux-musl"
QUERY_ENGINE_URL="https://binaries.prisma.sh/all_commits/${PRISMA_VERSION}/${OS}/query-engine.gz"
MIGRATION_ENGINE_URL="https://binaries.prisma.sh/all_commits/${PRISMA_VERSION}/${OS}/migration-engine.gz"

# download all engines and unzip them if they don't exist
if [ ! -f ./query-engine ]; then
  curl -L $QUERY_ENGINE_URL | gunzip > ./query-engine
  chmod +x ./query-engine
fi
if [ ! -f ./migration-engine ]; then
  curl -L $MIGRATION_ENGINE_URL | gunzip > ./migration-engine
  chmod +x ./migration-engine
fi