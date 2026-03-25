<div align="center">
    <pre style="background: none;">
   █████████     ███████    ███████████   █████████  
  ███░░░░░███  ███░░░░░███ ░█░░░███░░░█  ███░░░░░███ 
 ███     ░░░  ███     ░░███░   ░███  ░  ░███    ░███ 
░███         ░███      ░███    ░███     ░███████████ 
░███    █████░███      ░███    ░███     ░███░░░░░███ 
░░███  ░░███ ░░███     ███     ░███     ░███    ░███ 
 ░░█████████  ░░░███████░      █████    █████   █████
  ░░░░░░░░░     ░░░░░░░       ░░░░░    ░░░░░   ░░░░░ 
    </pre>
</div>

<div align="center">
    <h3>A collection of small, focused Go utility packages</h3>
    <h6>Currently under active development and breaking changes are possible</h6>
</div>

```
go get github.com/0x626f/gota
```

---

## Packages

### `bitflag` — bit-packed boolean flags

```go
const (
    Read  bitflag.BitFlag = 1 << iota
    Write
    Exec
)

var perms bitflag.BitFlag
perms.Add(Read | Write)
perms.Has(Read)   // true
perms.Has(Exec)   // false
perms.Delete(Write)
```

---

### `env` — populate structs from environment variables

Fields are mapped with the `env` struct tag. Nested structs are walked
recursively. Slices are populated from comma-separated values. Types
implementing `encoding.TextUnmarshaler` are handled automatically.

```go
type Config struct {
    Host    string        `env:"APP_HOST"`
    Port    int           `env:"APP_PORT"`
    Timeout time.Duration `env:"APP_TIMEOUT"`
    Tags    []string      `env:"APP_TAGS"`
}

var cfg Config
if err := env.Unmarshal(&cfg); err != nil {
    log.Fatal(err)
}
```

Helper functions:

```go
host := env.GetEnv("APP_HOST", "localhost") // returns default if unset
env.SetEnv("APP_HOST", "example.com")
```

---

### `event` — event routing and broadcast streams

**Router** — dispatches events to registered handlers by ID.

```go
router := event.NewRouter(event.RouterParams{
    Async: true,
    OnErr: func(e event.Event, err error) {
        log.Printf("event %s: %v", e.Id(), err)
    },
})

router.OnEvent("user.created", func(data any) error {
    user := data.(*User)
    return sendWelcomeEmail(user)
})

router.Route(myEvent)
```

**Stream** — broadcasts values from one source channel to many listeners.

```go
stream := event.NewStream[*Message](event.StreamParams{StreamSize: 32})

sub1 := stream.Listen()
sub2 := stream.Listen()

source := make(chan *Message)
stream.Bind(source)

// Both sub1 and sub2 receive every message sent to source.
```

---

### `workers` — background task utilities

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Run a task on a fixed interval.
workers.RegisterWorkerOnDelay(ctx, func() {
    syncCache()
}, 30*time.Second)

// Run a task whenever a signal arrives.
trigger := make(chan any)
workers.RegisterWorkerOnSignal(ctx, func() {
    flushMetrics()
}, trigger)

// Run a typed task for each value received.
jobs := make(chan Job)
workers.RegisterWorkerOnEvent(ctx, func(j Job) {
    process(j)
}, jobs)

// Retry a fallible operation.
err := workers.DoWithRetries(func() error {
    return connectToDB()
}, 3)

// Measure execution time.
duration, err := workers.DoWithStopwatch(func() error {
    return runMigration()
})
```

---

### `json` / `yaml` / `toml` — codec helpers

Thin wrappers that expose a consistent `Marshall` / `Unmarshal` pair across
formats. `json.Marshall` uses two-space indentation for human-readable output.

```go
data, err := json.Marshall(v)
err = json.Unmarshal(data, &v)

data, err = yaml.Marshall(v)
err = yaml.Unmarshal(data, &v)

data, err = toml.Marshall(v)
err = toml.Unmarshal(data, &v)
```
