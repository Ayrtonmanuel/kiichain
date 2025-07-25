#!/bin/bash
set -eux

# User balance of stake tokens
USER_COINS="100000000000stake"
# Amount of stake tokens staked
STAKE="100000000stake"
# Node IP address
NODE_IP="127.0.0.1"

# Home directory
HOME_DIR="$HOME"

# Validator moniker
MONIKER="coordinator"

# Validator directory
PROV_NODE_DIR=${HOME_DIR}/provider-${MONIKER}

# Coordinator key
PROV_KEY=${MONIKER}-key


# Clean start
pkill -f kiichaind &> /dev/null || true
rm -rf ${PROV_NODE_DIR}

# Build file and node directory structure
kiichaind init $MONIKER --chain-id provider --home ${PROV_NODE_DIR}
    jq ".app_state.gov.voting_params.voting_period = \"20s\" | .app_state.staking.params.unbonding_time = \"86400s\"" \
   ${PROV_NODE_DIR}/config/genesis.json > \
   ${PROV_NODE_DIR}/edited_genesis.json && mv ${PROV_NODE_DIR}/edited_genesis.json ${PROV_NODE_DIR}/config/genesis.json

sleep 1

# Create account keypair
kiichaind keys add $PROV_KEY --home ${PROV_NODE_DIR} --keyring-backend test --output json > ${PROV_NODE_DIR}/${PROV_KEY}.json 2>&1
sleep 1

# Add stake to user
PROV_ACCOUNT_ADDR=$(jq -r '.address' ${PROV_NODE_DIR}/${PROV_KEY}.json)
kiichaind genesis add-genesis-account $PROV_ACCOUNT_ADDR $USER_COINS --home ${PROV_NODE_DIR} --keyring-backend test
sleep 1


# Stake 1/1000 user's coins
kiichaind genesis gentx $PROV_KEY $STAKE --chain-id provider --home ${PROV_NODE_DIR} --keyring-backend test --moniker $MONIKER
sleep 1

kiichaind genesis collect-gentxs --home ${PROV_NODE_DIR} --gentx-dir ${PROV_NODE_DIR}/config/gentx/
sleep 1

sed -i -r "/node =/ s/= .*/= \"tcp:\/\/${NODE_IP}:26658\"/" ${PROV_NODE_DIR}/config/client.toml
sed -i -r 's/minimum-gas-prices = ""/minimum-gas-prices = "0stake"/g' ${PROV_NODE_DIR}/config/app.toml


# Start Kiichain
kiichaind start \
    --home ${PROV_NODE_DIR} \
    --rpc.laddr tcp://${NODE_IP}:26658 \
    --grpc.address ${NODE_IP}:9091 \
    --address tcp://${NODE_IP}:26655 \
    --p2p.laddr tcp://${NODE_IP}:26656 \
    --grpc-web.enable=false &> ${PROV_NODE_DIR}/logs
