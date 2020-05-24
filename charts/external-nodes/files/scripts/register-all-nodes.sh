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

DIR=$(dirname $0)
NODES_DEF_DIR=${DIR}/../nodes
REGISTER_NODE=${DIR}/register-node.sh

echo "Registering docker nodes"

for ii in $(ls ${NODES_DEF_DIR}/node*.env)
do
  echo "
-----------------
Processing $ii
"
set -a
source $ii
set +a
$REGISTER_NODE

done


