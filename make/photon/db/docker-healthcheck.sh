#!/bin/bash
set -eo pipefail

h="$(hostname -i || echo '127.0.0.1')"
host="${h%%[[:space:]]*}" #remove the trailing space
user="${POSTGRES_USER:-postgres}"
db="${POSTGRES_DB:-$POSTGRES_USER}"
export PGPASSWORD="${POSTGRES_PASSWORD:-}"

args=(
	# force postgres to not use the local unix socket (test "external" connectibility)
	--host "$host"
	--username "$user"
	--dbname "$db"
	--quiet --no-align --tuples-only
)

if select="$(echo 'SELECT 1' | psql "${args[@]}")" && [ "$select" = '1' ]; then
	exit 0
fi

exit 1