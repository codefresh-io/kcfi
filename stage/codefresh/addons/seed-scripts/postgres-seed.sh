#!/bin/bash

set -xeuo pipefail

POSTGRES_DATABASE="${POSTGRES_DATABASE:-codefresh}"
POSTGRES_AUDIT_DATABASE="${POSTGRES_AUDIT_DATABASE:-audit}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"

# To create a separate non-privileged user the for Codefresh,
# which has access only to the relevant databases, it is needed to specify 
# additionally the POSTGRES_SEED_USER and POSTGRES_SEED_PASSWORD vars.
# Otherwise only POSTGRES_USER and POSTGRES_PASSWORD will be used both
# during seed job execution and runtime

POSTGRES_SEED_USER="${POSTGRES_SEED_USER:-$POSTGRES_USER}"
POSTGRES_SEED_PASSWORD="${POSTGRES_SEED_PASSWORD:-$POSTGRES_PASSWORD}"

function createDB() {
    psql \
        --host ${POSTGRES_HOST} \
        --port ${POSTGRES_PORT} \
        -U ${POSTGRES_SEED_USER} \
        -c \
        "create database ${POSTGRES_DATABASE}"
}

function createAuditDB() {
    psql \
        --host ${POSTGRES_HOST} \
        --port ${POSTGRES_PORT} \
        -U ${POSTGRES_SEED_USER} \
        -c \
        "create database ${POSTGRES_AUDIT_DATABASE}"    
}
 
function createUser() {
    echo "Creating a separate non-privileged user for Codefresh"
    psql \
        --host ${POSTGRES_HOST} \
        --port ${POSTGRES_PORT} \
        -U ${POSTGRES_SEED_USER} \
        -c "CREATE USER ${POSTGRES_USER} WITH PASSWORD '${POSTGRES_PASSWORD}'"
}

function grantPrivileges() {
    psql \
        --host ${POSTGRES_HOST} \
        --port ${POSTGRES_PORT} \
        -U ${POSTGRES_SEED_USER} \
        -c "GRANT ALL ON DATABASE ${POSTGRES_DATABASE} TO ${POSTGRES_USER}"
}

function grantAuditPrivileges() {
    psql \
        --host ${POSTGRES_HOST} \
        --port ${POSTGRES_PORT} \
        -U ${POSTGRES_SEED_USER} \
        -c "GRANT ALL ON DATABASE ${POSTGRES_AUDIT_DATABASE} TO ${POSTGRES_USER}"
}

function runSeed() {

    export PGPASSWORD=${POSTGRES_SEED_PASSWORD}

    createDB
    createAuditDB

    if [[ "${POSTGRES_SEED_USER}" != "${POSTGRES_USER}" ]]; then
        createUser
    else   
        echo "There is no a separate user specified for the seed job, skipping user creation"
    fi

    grantPrivileges
    grantAuditPrivileges
}

runSeed