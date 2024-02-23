set -x

function testSchellar() {

    docker-compose -f ../docker-compose.test.yml up -d

    sleep 5
    export BACKEND="postgres"

    export POSTGRES_HOST=127.0.0.1
    export POSTGRES_PORT=6432
    export POSTGRES_DB=schellar_test
    export POSTGRES_USER=postgres
    export POSTGRES_PASSWORD=postgres
    export POSTGRES_MIGRATIONS_DIR="$(pwd)/migrations"
    go test -run Integration ./...
    docker-compose -f ../docker-compose.test.yml down

}


dirname="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${dirname}

cd schellar
go test -v -short ./...

trap testSchellar err exit
