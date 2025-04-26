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

# Using with Custom Loggers

The library uses its own Logger interface but provides utilities to adapt your own logger:

1. For loggers with Debug, Info, Warn, and Error methods, use the adapter:

```go
import (
    p2p_database "github.com/dTelecom/p2p-realtime-database"
)

// Adapt your simple logger to the required interface
adaptedLogger := p2p_database.NewLoggerAdapter(yourLogger)

// Use the adapted logger with Connect
db, err := p2p_database.Connect(ctx, config, adaptedLogger)
```

2. To use the built-in console logger:

```go
logger := p2p_database.NewConsoleLogger()
db, err := p2p_database.Connect(ctx, config, logger)
```

The SimpleLogger interface required by NewLoggerAdapter is:

```go
type SimpleLogger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
}
```