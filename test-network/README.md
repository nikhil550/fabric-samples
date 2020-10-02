# Adding ordering nodes

This repo contains two scripts that add an ordering node for to your ordering service.

`ordererUpdate.sh` adds a second ordering node that belongs to the ordering organization.

`org1OrdererUpdate.sh` adds an ordering node that belongs to org1.

To use the script you need to create a network using the CAs:
```
./network.sh up -ca
```

Then create an application channel:
```
./network.sh createChannel
```

Then use the script to register the second ordering node and add the script to the
system channel and the application channel. Put the system channel first
```
ordererUpdate.sh register system-channel mychannel
```

You can also use the same command to add the Org1 ordering node.
```
org1OrdererUpdate.sh register system-channel mychannel
```

## testing

You can use the `deployCC` command to test your new ordering node. Simply open
the `network.sh` script and comment out the script that deploys the CC using the
first ordering node, and uncomment the one that uses the second ordering organization
node or the Org1 ordering node.

```
function deployCC() {

  scripts/deployCC.sh $CHANNEL_NAME $CC_RUNTIME_LANGUAGE $VERSION $CLI_DELAY $MAX_RETRY $VERBOSE

#  scripts/deployCCorderer2.sh $CHANNEL_NAME $CC_RUNTIME_LANGUAGE $VERSION $CLI_DELAY $MAX_RETRY $VERBOSE

#  scripts/deployCCorg1.sh $CHANNEL_NAME $CC_RUNTIME_LANGUAGE $VERSION $CLI_DELAY $MAX_RETRY $VERBOSE
```
