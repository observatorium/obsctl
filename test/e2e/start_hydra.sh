#!/bin/bash

set -e
set -o pipefail

trap 'kill 0' SIGTERM

OS="$(uname)"
case $OS in
'Linux')
    HYDRA_OS='linux'
    HYDRA_CONFIG='hydra.yaml'
    ;;
'Darwin')
    HYDRA_OS='macos'
    HYDRA_CONFIG='hydra-for-macOS.yaml'
    ;;
*)
    echo "Unsupported OS for this script: $OS"
    exit 1
    ;;
esac

WD="$(pwd)"
gethydra() {
    mkdir -p mkdir -p $WD/tmp/bin
    echo "-------------------------------------------"
    echo "- Downloading ORY Hydra...  -"
    echo "-------------------------------------------"
    curl -L "https://github.com/ory/hydra/releases/download/v1.9.1/hydra_1.9.1-sqlite_${HYDRA_OS}_64bit.tar.gz" | tar -xzf - -C $WD/tmp/bin hydra
}
startHydra() {
    (DSN=memory $WD/tmp/bin/hydra serve all --dangerous-force-http --config $WD/$HYDRA_CONFIG &>/dev/null) &
    echo "-------------------------------------------"
    echo "- Waiting for Hydra to come up...  -"
    echo "-------------------------------------------"
    until curl --output /dev/null --silent --fail --insecure http://127.0.0.1:4444/.well-known/openid-configuration; do
        printf '.'
        sleep 1
    done
    echo ""
}

gethydra
startHydra
exit 0
