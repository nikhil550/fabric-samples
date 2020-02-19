
## Set environment

export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="OrdererMSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/ordererOrganizations/example.com/users/Admin@example.com/msp
export ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# IF you are using a CA, Register the new orderer and create orderer MSP and TLS certificates

function registerOrderer() {
  export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org1.example.com

  fabric-ca-client register --caname ca-org1 --id.name orderer --id.secret ordererpw --id.type orderer --tls.certfiles ${PWD}/organizations/fabric-ca/org1/tls-cert.pem

  fabric-ca-client enroll -u https://orderer:ordererpw@localhost:7054 --caname ca-org1 -M ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/msp --csr.hosts orderer.org1.example.com --tls.certfiles ${PWD}/organizations/fabric-ca/org1/tls-cert.pem

  fabric-ca-client enroll -u https://orderer:ordererpw@localhost:7054 --caname ca-org1 -M ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls --enrollment.profile tls --csr.hosts orderer.org1.example.com --csr.hosts localhost --tls.certfiles ${PWD}/organizations/fabric-ca/org1/tls-cert.pem

  echo 'NodeOUs:
    Enable: true
    ClientOUIdentifier:
      Certificate: cacerts/localhost-7054-ca-org1.pem
      OrganizationalUnitIdentifier: client
    PeerOUIdentifier:
      Certificate: cacerts/localhost-7054-ca-org1.pem
      OrganizationalUnitIdentifier: peer
    AdminOUIdentifier:
      Certificate: cacerts/localhost-7054-ca-org1.pem
      OrganizationalUnitIdentifier: admin
    OrdererOUIdentifier:
      Certificate: cacerts/localhost-7054-ca-org1.pem
      OrganizationalUnitIdentifier: orderer' > ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/msp/config.yaml

    # Move certs around to make them easier to mount and point to with env variables

  cp ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/tlscacerts/* ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/ca.crt
  cp ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/signcerts/* ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/server.crt
  cp ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/keystore/* ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/server.key

  mkdir ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/msp/tlscacerts
  cp ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/msp/tlscacerts/* ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

}
  # Pull the system channel config and decode it

function updateConfig() {

  export FLAG=$(if [ "$(uname -s)" == "Linux" ]; then echo "-w 0"; else echo "-b 0"; fi)
  export TLS_CERT=$(cat ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/server.crt | base64 $FLAG)

  export CA_ROOT_CERT=$(cat ${PWD}/organizations/peerOrganizations/org1.example.com/msp/cacerts/localhost-7054-ca-org1.pem | base64 $FLAG)

  export TLS_ROOT_CERT=$(cat ${PWD}/organizations/peerOrganizations/org1.example.com/orderers/orderer.org1.example.com/tls/tlscacerts/tls-localhost-7054-ca-org1.pem | base64 $FLAG)

  MSPID="Org1MSP"

  mkdir config

  peer channel fetch config config/config_block.pb -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com -c $CHANNEL_NAME --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

  cd config

  configtxlator proto_decode --input config_block.pb --type common.Block --output config_block.json
  jq .data.data[0].payload.data.config config_block.json > config.json
  cp config.json modified_config.json

  # Who are the current current consenters

  CONSENTER_LIST=$(cat modified_config.json | jq .channel_group.groups.Orderer.values.ConsensusType.value.metadata.consenters | sed 's/[][]//g')

  echo "$CONSENTER_LIST"

  ## Add new consenter to the list in the confiuration

  echo '[
  '$CONSENTER_LIST',
    {
      "client_tls_cert": "'"$TLS_CERT"'",
      "host": "orderer.org1.example.com",
      "port": 6050,
      "server_tls_cert": "'"$TLS_CERT"'"
    }
  ]' > new_consenter.json


  jq -s '.[0] * {"channel_group":{"groups":{"Orderer":{"values":{"ConsensusType":{"value":{"metadata":{consenters:.[1]}}}}}}}}' config.json new_consenter.json > second_config.json

  set -x
  cat second_config.json | jq .channel_group.groups.Orderer.values.ConsensusType.value.metadata.consenters
  set +x

  ## Add new orderer address to the list in the confiuration

  ORDERER_ADRESSES=$(cat second_config.json | jq .channel_group.values.OrdererAddresses.value.addresses | sed 's/[][]//g')

  echo "$ORDERER_ADRESSES"

  echo '[
    '$ORDERER_ADRESSES',
    "orderer.org1.example.com:6050"
  ]' > new_addresses.json

  jq -s '.[0] * {"channel_group":{"values":{"OrdererAddresses":{"value":{addresses:.[1]}}}}}' second_config.json new_addresses.json > third_config.json


  # Add new org to the orderers


  ## Create new org definition

  ./../createOrgDef.sh $MSPID $CA_ROOT_CERT $TLS_ROOT_CERT

  set -x
  jq -s '.[0] * {"channel_group":{"groups":{"Orderer":{"groups": {"'"$MSPID"'":.[1]}}}}}' third_config.json Org1Def.json > modified_config.json
  set +x

  ## Check the new config

  set -x
  cat modified_config.json | jq .channel_group.groups.Orderer.values.ConsensusType.value.metadata.consenters
  set +x

  set -x
  cat modified_config.json | jq .channel_group.values.OrdererAddresses.value.addresses
  set +x

  set -x
  cat modified_config.json | jq .channel_group.groups.Orderer.groups.${MSPID}
  set +x

  # Create the coniguration update

  configtxlator proto_encode --input config.json --type common.Config --output config.pb
  configtxlator proto_encode --input modified_config.json --type common.Config --output modified_config.pb
  configtxlator compute_update --channel_id $CHANNEL_NAME --original config.pb --updated modified_config.pb --output config_update.pb

  configtxlator proto_decode --input config_update.pb --type common.ConfigUpdate --output config_update.json
  echo '{"payload":{"header":{"channel_header":{"channel_id":"'$CHANNEL_NAME'", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' | jq . > config_update_in_envelope.json
  configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope --output config_update_in_envelope.pb

  # Submit the update

  cd ..

  peer channel update -f config/config_update_in_envelope.pb -c $CHANNEL_NAME -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA

  # wait for the updated block to happen

  sleep 5

  if [ $CHANNEL_NAME == "system-channel" ]; then

    # Fetch new system channel configuration and copy it

    peer channel fetch config config/new_config_block.pb -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com -c $CHANNEL_NAME --tls --cafile $ORDERER_CA
    cp config/new_config_block.pb system-genesis-block/config.block

    # create the orderer
    docker-compose -f docker/docker-compose-test-ordererOrg1.yaml up -d 2>&1

  fi

  rm -rf config/*

}

if [[ $# -ge 1 ]] ; then
  key="$1"
  if [[ "$key" == "register" ]]; then
      registerOrderer
      shift
  fi
  if [[ "$key" == "down" ]]; then
      docker-compose -f docker/docker-compose-test-ordererOrg1.yaml down --volumes --remove-orphans
      ./ordererUpdate.sh down
      ./network.sh down
      shift
  fi
fi


while [[ $# -ge 1 ]] ; do
  export CHANNEL_NAME="$1"
  updateConfig
  shift
done
