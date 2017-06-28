#! /usr/bin/env bash
set -e

# This scripts rebuilds the latest.sql file included in the schema package.

gb generate github.com/openbankit/horizon/db2/schema
gb build
dropdb horizon_schema --if-exists
createdb horizon_schema
DATABASE_URL=postgres://localhost/horizon_schema?sslmode=disable ./bin/horizon db migrate up

pg_dump postgres://localhost/horizon_schema?sslmode=disable --schema=public --no-owner --no-acl --inserts > src/github.com/openbankit/horizon/db2/schema/latest.sql
pg_dump postgres://localhost/horizon_schema?sslmode=disable --clean --if-exists --no-owner --no-acl --inserts > src/github.com/openbankit/horizon/test/scenarios/blank-horizon.sql

gb generate github.com/openbankit/horizon/db2/schema
gb generate github.com/openbankit/horizon/test
gb build
