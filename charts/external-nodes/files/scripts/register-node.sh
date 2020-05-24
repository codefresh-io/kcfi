#!/usr/bin/env bash
#

usage() {
    echo "Register docker-node in consul
    $0 <node-addres> [port]
    "
}

print_envs() {
  for ii in $@
  do
    echo "${ii}=${!ii}"
  done
}

set -e
NODE_ADDRESS=${1:-$NODE_ADDRESS}
if [[ -z "${NODE_ADDRESS}" ]]; then
  echo "Error: no node address provided"
  usage
  exit 1
fi

NODE_PORT=${2:-${NODE_PORT:-2376}}
NODE_NAME=${NODE_NAME:-${NODE_ADDRESS}}
CONSUL=${CONSUL:-"http://cf-consul:8500"}
NODE_CLUSTER=${NODE_CLUSTER:-"codefresh"}
NODE_ROLE=${NODE_ROLE:-builder}

print_envs NODE_ADDRESS NODE_PORT NODE_NAME CONSUL NODE_CLUSTER NODE_ROLE DRY_RUN

PROVIDER='
{ 
    "name": "remote", 
    "type": "customer" 
}'
SYSTEM_DATA='{"os_name": "linux"}'
NODE_SERVICE_DEF='
{
    "Node": "'${NODE_NAME}'",
    "Address": "'${NODE_ADDRESS}'",
    "Service": {
        "Service": "docker-node",
        "Tags": [
            "dind",
            "noagent",
            "account_codefresh",
            "type_builder"
        ],
        "Address": "'${NODE_ADDRESS}'",
        "Port": '${NODE_PORT}'
    },
    "Check": {
        "Node": "",
        "CheckID": "service:docker-node",
        "Name": "Remote Node Check",
        "Notes": "Check builder is up and running",
        "Output": "Builder alive and reachable",
        "Status": "passing",
        "ServiceID": "docker-node"
    }
}'
echo "Registering node $NODE_NAME in consul. Configuration: ${NODE_SERVICE}"

if [[ -n ${DRY_RUN} ]]; then
    print_envs NODE_SERVICE_DEF
    echo "DRY RUN MODE - only printing envs and exit"
    exit
fi
curl -X PUT -d "${NODE_SERVICE_DEF}" ${CONSUL}/v1/catalog/register
curl -X PUT -d "${NODE_ADDRESS}" ${CONSUL}/v1/kv/services/docker-node/${NODE_NAME}/publicAddress
curl -X PUT -d "${NODE_CLUSTER}" ${CONSUL}/v1/kv/services/docker-node/${NODE_NAME}/account
curl -X PUT -d "${NODE_ROLE}" ${CONSUL}/v1/kv/services/docker-node/${NODE_NAME}/role
curl -X PUT -d "${PROVIDER}" ${CONSUL}/v1/kv/services/docker-node/${NODE_NAME}/systemData
curl -X PUT -d "${SYSTEM_DATA}" ${CONSUL}/v1/kv/services/docker-node/${NODE_NAME}/provider
