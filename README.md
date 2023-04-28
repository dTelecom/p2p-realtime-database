# Example run
```
go build examples/main.go

export ETHEREUM_NETWORK_HOST=https://arbitrum-goerli.infura.io/v3/
export ETHEREUM_NETWORK_KEY=47d2344cdbaa4c89a395fea69d452261
export ETHEREUM_CONTRACT_ADDRESS=0xa99885B0Ce9cE7C8B93c15bFc4deeAec419f2393

./main -pk <privateKey> #error logging mode
./main -pk <privateKey> -v #warning logging mode
./main -pk <privateKey> -vv #info logging mode
./main -pk <privateKey> -vvv #debug logging mode
```

# Usage
```
> debug on #debug mode
> debug on info
> debug on warning
> debug on error
> debug off
> set key value
> get key value
> del key
> list
> subscribe <event>
> publish <event> <value>
```