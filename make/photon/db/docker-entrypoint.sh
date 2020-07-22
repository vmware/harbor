#!/bin/bash
set -e

# usage: file_env VAR [DEFAULT]
#    ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
        local var="$1"
        local fileVar="${var}_FILE"
        local def="${2:-}"
        if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
                echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
                exit 1
        fi
        local val="$def"
        if [ "${!var:-}" ]; then
                val="${!var}"
        elif [ "${!fileVar:-}" ]; then
                val="$(< "${!fileVar}")"
        fi
        export "$var"="$val"
        unset "$fileVar"
}

# look specifically for PG_VERSION, as it is expected in the DB dir
if [ ! -s "$PGDATA/PG_VERSION" ]; then
        file_env 'POSTGRES_INITDB_ARGS'
        if [ "$POSTGRES_INITDB_XLOGDIR" ]; then
                export POSTGRES_INITDB_ARGS="$POSTGRES_INITDB_ARGS --xlogdir $POSTGRES_INITDB_XLOGDIR"
        fi
        initdb -D $PGDATA  -U postgres -E UTF-8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 $POSTGRES_INITDB_ARGS
        # check password first so we can output the warning before postgres
        # messes it up
        file_env 'POSTGRES_PASSWORD'
        if [ "$POSTGRES_PASSWORD" ]; then
                pass="PASSWORD '$POSTGRES_PASSWORD'"
                authMethod=md5
        else
                # The - option suppresses leading tabs but *not* spaces. :)
                cat >&2 <<-EOF
                        ****************************************************
                        WARNING: No password has been set for the database.
                                        This will allow anyone with access to the
                                        Postgres port to access your database. In
                                        Docker's default configuration, this is
                                        effectively any other container on the same
                                        system.
                                        Use "-e POSTGRES_PASSWORD=password" to set
                                        it in "docker run".
                        ****************************************************
EOF

                pass=
                authMethod=trust
        fi

        {
                echo
                echo "host all all all $authMethod"
        } >> "$PGDATA/pg_hba.conf"
        echo `whoami`
        # internal start of server in order to allow set-up using psql-client
        # does not listen on external TCP/IP and waits until start finishes
        pg_ctl -D "$PGDATA" -o "-c listen_addresses=''" -w start

        file_env 'POSTGRES_USER' 'postgres'
        file_env 'POSTGRES_DB' "$POSTGRES_USER"

        psql=( psql -v ON_ERROR_STOP=1 )

        if [ "$POSTGRES_DB" != 'postgres' ]; then
                "${psql[@]}" --username postgres <<-EOSQL
                        CREATE DATABASE "$POSTGRES_DB" ;
EOSQL
                echo
        fi

        if [ "$POSTGRES_USER" = 'postgres' ]; then
                op='ALTER'
        else
                op='CREATE'
        fi
        "${psql[@]}" --username postgres <<-EOSQL
                $op USER "$POSTGRES_USER" WITH SUPERUSER $pass ;
EOSQL
        echo

        psql+=( --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" )

        echo
        for f in /docker-entrypoint-initdb.d/*; do
                case "$f" in
                        *.sh)     echo "$0: running $f"; . "$f" ;;
                        *.sql)    echo "$0: running $f"; "${psql[@]}" -f "$f"; echo ;;
                        *.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
                        *)        echo "$0: ignoring $f" ;;
                esac
                echo
        done

        PGUSER="${PGUSER:-postgres}" \
        pg_ctl -D "$PGDATA" -m fast -w stop

        echo
        echo 'PostgreSQL init process complete; ready for start up.'
        echo
fi

POSTGRES_PARAMETER=''

file_env 'POSTGRES_MAX_CONNECTIONS' '1024'
# The max value of 'max_connections' is 262143
if [ $POSTGRES_MAX_CONNECTIONS -le 0 ] || [ $POSTGRES_MAX_CONNECTIONS -gt 262143 ]; then
        POSTGRES_MAX_CONNECTIONS=262143
fi

POSTGRES_PARAMETER="${POSTGRES_PARAMETER} -c max_connections=${POSTGRES_MAX_CONNECTIONS}"

postgres -D $PGDATA $POSTGRES_PARAMETER
