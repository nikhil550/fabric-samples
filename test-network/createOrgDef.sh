MSPID="$1"
CA_CERT="$2"
TLS_ROOT_CERT="$3"

echo '{
  "groups": {},
  "mod_policy": "Admins",
  "policies": {
    "Admins": {
      "mod_policy": "Admins",
      "policy": {
        "type": 1,
        "value": {
          "identities": [
            {
              "principal": {
                "msp_identifier": "'"$MSPID"'",
                "role": "ADMIN"
              },
              "principal_classification": "ROLE"
            }
          ],
          "rule": {
            "n_out_of": {
              "n": 1,
              "rules": [
                {
                  "signed_by": 0
                }
              ]
            }
          },
          "version": 0
        }
      },
      "version": "0"
    },
    "Readers": {
      "mod_policy": "Admins",
      "policy": {
        "type": 1,
        "value": {
          "identities": [
            {
              "principal": {
                "msp_identifier": "'"$MSPID"'",
                "role": "MEMBER"
              },
              "principal_classification": "ROLE"
            }
          ],
          "rule": {
            "n_out_of": {
              "n": 1,
              "rules": [
                {
                  "signed_by": 0
                }
              ]
            }
          },
          "version": 0
        }
      },
      "version": "0"
    },
    "Writers": {
      "mod_policy": "Admins",
      "policy": {
        "type": 1,
        "value": {
          "identities": [
            {
              "principal": {
                "msp_identifier": "'"$MSPID"'",
                "role": "MEMBER"
              },
              "principal_classification": "ROLE"
            }
          ],
          "rule": {
            "n_out_of": {
              "n": 1,
              "rules": [
                {
                  "signed_by": 0
                }
              ]
            }
          },
          "version": 0
        }
      },
      "version": "0"
    }
  },
  "values": {
    "MSP": {
      "mod_policy": "Admins",
      "value": {
        "config": {
          "admins": [],
          "crypto_config": {
            "identity_identifier_hash_function": "SHA256",
            "signature_hash_family": "SHA2"
          },
          "fabric_node_ous": {
            "admin_ou_identifier": {
              "certificate": "'"$CA_CERT"'",
              "organizational_unit_identifier": "admin"
            },
            "client_ou_identifier": {
              "certificate": "'"$CA_CERT"'",
              "organizational_unit_identifier": "client"
            },
            "enable": true,
            "orderer_ou_identifier": {
              "certificate": "'"$CA_CERT"'",
              "organizational_unit_identifier": "orderer"
            },
            "peer_ou_identifier": {
              "certificate": "'"$CA_CERT"'",
              "organizational_unit_identifier": "peer"
            }
          },
          "intermediate_certs": [],
          "name": "'"$MSPID"'",
          "organizational_unit_identifiers": [],
          "revocation_list": [],
          "root_certs": [
            "'"$CA_CERT"'"
          ],
          "signing_identity": null,
          "tls_intermediate_certs": [],
          "tls_root_certs": [
            "'"$TLS_ROOT_CERT"'"
          ]
        },
        "type": 0
      },
      "version": "0"
    }
  },
  "version": "0"
}' > Org1Def.json
