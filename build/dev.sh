#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
mkdir -p $DIR/../target/ganache-data
cd $DIR/../contracts
ganache-cli -m 'refuse result toy bunker royal small story exhaust know piano base stand' --db "$DIR/../target/ganache-data/" -d -i 15 &
G_PID="$!"

while :; do
    if nc -z localhost 8545; then
        break
    fi
    kill -0 "$G_PID" || {
        echo "Error: ganache crashed. Check the logs and try again?"
        exit 1
    }
    sleep 1
done

truffle migrate --network local
echo "Done migrating contracts."
kill $G_PID
echo "Restarting Ganache with blocktime..."
set -E
trap "kill 0" SIGINT

cyan=6; green=2; yellow=3; blue=4; pink=5;

prefix_cmd () {
    {
        eval "$3"
    } 2>&1 | sed -e "s/\(.*\)/$(tput setaf $2)[$1] \1$(tput sgr0)/"
}

(
    prefix_cmd 'ganache' ${cyan} \
        "ganache-cli -m 'refuse result toy bunker royal small story exhaust know piano base stand' --db '$DIR/../target/ganache-data/' -b 1 -i 15 -d"
) &

sleep 3

(
    prefix_cmd 'plasma' ${pink} "$DIR/../target/plasma --config $DIR/config-local.yaml start-root"
) &

wait