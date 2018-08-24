#!/usr/bin/env bash

DATA_DIR=$1
GETH_DATA_DIR="${DATA_DIR}/chain"

FIRST_USER_SUFFIX=$(printf '\n\n' | geth --datadir "${GETH_DATA_DIR}" account new  | grep "Address" | awk '{ printf("%s\n", substr($2, 2, 40)) }')

KEYSTORE_DIR="${GETH_DATA_DIR}/keystore"
KEYSTORE_FILE_FILTER="${KEYSTORE_DIR}/*${FIRST_USER_SUFFIX}"
KEYSTORE_FILE="$(ls -l ${KEYSTORE_FILE_FILTER} | cut -d '/' -f 2-25)"
echo "Keystore file: ${KEYSTORE_FILE}"
FIRST_USER_ADDRESS="0x${FIRST_USER_SUFFIX}"
echo "First user address: ${FIRST_USER_ADDRESS}"

SECOND_USER_SUFFIX=$(printf '\n\n' | geth --datadir "${GETH_DATA_DIR}" account new  | grep "Address" | awk '{ printf("%s\n", substr($2, 2, 40)) }')
SECOND_USER_ADDRESS="0x${SECOND_USER_SUFFIX}"
echo "Second user address: ${SECOND_USER_ADDRESS}"

geth --dev --dev.period 6 --identity "LocalPlasma" --rpc --rpccorsdomain "*" --datadir "${GETH_DATA_DIR}" --port "30303" --nodiscover --rpcapi "db,eth,net,web3" --networkid 15 --nat "any" 2>$DATA_DIR/geth.log &
GETH_PID=$!
sleep 3

pushd contracts

TRUFFLE_OUT="$(truffle migrate --network local --reset 2>&1 )"
PQ_ADDRESS=$(echo "${TRUFFLE_OUT}" | grep 'PriorityQueue\:' | cut -d ':' -f 2 | cut -d ' ' -f 2)
echo "PriorityQueue address: ${PQ_ADDRESS}"
PLASMA_ADDRESS=$(echo "${TRUFFLE_OUT}" | grep 'Plasma\:' | cut -d ':' -f 2 | cut -d ' ' -f 2)
echo "Plasma address: ${PLASMA_ADDRESS}"

popd

OPTIONS="{\"database_dir\": \"${DATA_DIR}/plasma\", \"plasma_address\": \"${PLASMA_ADDRESS}\", \"pq_address\": \"${PQ_ADDRESS}\", \"user_address\": \"${FIRST_USER_ADDRESS}\", \"keystore_dir\": \"${KEYSTORE_DIR}\", \"keystore_file\": \"${KEYSTORE_FILE}\"}"
echo "${OPTIONS}" | mustache ./mustache/test-config.mustache | tee $DATA_DIR/test-config.yaml

# Unlock accounts
UNLOCK=$(geth --exec 'var i; for (i=0; i < eth.accounts.length; i++) { personal.unlockAccount(eth.accounts[i], "", 0) }' attach ipc:"${GETH_DATA_DIR}/geth.ipc")

plasma --config $DATA_DIR/test-config.yaml start &> $DATA_DIR/plasma.log &
PLASMA_PID=$!
echo "Started plasma node, pid is ${PLASMA_PID}"
sleep 10

#plasma --config $DATA_DIR/test-config.yaml deposit --amount 200000000
#
#plasma --config $DATA_DIR/test-config.yaml send --to ${SECOND_USER_ADDRESS} --amount 50000000

kill $PLASMA_PID
sleep 1
kill $GETH_PID