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

The library uses its own Logger interface defined in the `internal/common` package. If you have your own logger, there are a few ways to use it with this library:

1. If your logger implements all the required methods, you can use it directly with `Connect`.

2. If your logger has a simpler interface (with just Debug, Info, Warn, Error methods), you can use the adapter:

```go
import (
    p2p_database "github.com/dTelecom/p2p-realtime-database"
    "github.com/dTelecom/p2p-realtime-database/internal/common"
)

// Adapt your simple logger to the required interface
adaptedLogger := common.NewLoggerAdapter(yourLogger)

// Use the adapted logger with Connect
db, err := p2p_database.Connect(ctx, config, adaptedLogger)
```

3. You can also use the console logger included in the library:

```go
logger := new(common.ConsoleLogger)
db, err := p2p_database.Connect(ctx, config, logger)
```