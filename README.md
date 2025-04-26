# Example run
```
go build examples/main.go

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
> subscribe <event>
> publish <event> <value>
> peers
```