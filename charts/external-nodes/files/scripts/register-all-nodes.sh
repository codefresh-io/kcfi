#!/usr/bin/env bash
#

reset_node_vars() {
    unset NODE_ADDRESS
    unset NODE_PORT
    unset NODE_NAME
    unset CONSUL
    unset NODE_CLUSTER
    unset NODE_ROLE
}

echo "-----------------------------
Starting $0 at $(date)
"

DIR=$(dirname $0)
NODES_DEF_DIR=${DIR}/../nodes
REGISTER_NODE=${DIR}/register-node.sh

echo "Registering docker nodes"
FAILED_NODES=
for ii in $(ls ${NODES_DEF_DIR}/*.env)
do
  echo "
-----------------
Processing $ii
"
reset_node_vars
set -a
source $ii
set +a
$REGISTER_NODE
if [[ $? != 0 ]]; then
    echo "ERROR: NODE $NODE_NAME registration failed"
    FAILED_NODES+="$NODE_NAME "
fi
done

if [[ -n "${FAILED_NODES}" ]]; then
  echo "FAILED NODES: $FAILED_NODES"
  exit 1
fi


