set -x
dirname="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${dirname}

cd schellar
go test -v -short ./...

docker-compose -f ../docker-compose.test.yml up -d

trap 'docker-compose -f ../docker-compose.test.yml down' err exit

sleep 5
export BACKEND="postgres"
export POSTGRES_DATABASE_URL="host=127.0.0.1 port=6432 user=postgres password=postgres database=schellar_test"
export POSTGRES_MIGRATIONS_DIR="$(pwd)/migrations"
export MONGO_ADDRESS=127.0.0.1
export MONGO_USERNAME=root
export MONGO_PASSWORD=root
export MONGO_DB=admin
go test -run Integration ./...
