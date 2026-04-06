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

Fields are mapped with the `env` struct tag. Use `default` to supply a
fallback value when the variable is absent. Nested structs are walked
recursively. Slices are populated from comma-separated values. Types
implementing `encoding.TextUnmarshaler` are handled automatically.

```go
type Config struct {
    Host    string        `env:"APP_HOST"    default:"localhost"`
    Port    int           `env:"APP_PORT"    default:"8080"`
    Timeout time.Duration `env:"APP_TIMEOUT" default:"30s"`
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

Workers are created with a `New*` constructor and started with `.Run()`.
`.Run()` is idempotent — calling it more than once has no effect. Tasks use
`func() error`; errors and panics are forwarded to optional handlers.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Run a task immediately, then on every tick.
w := workers.NewWorkerOnTicker(ctx, func() error {
    return syncCache()
}, 30*time.Second)
w.OnError(func(err error) { log.Println("sync:", err) })
w.Run()

// Run a task each time a signal is received on the channel.
trigger := make(chan struct{})
w = workers.NewWorkerOnSignal(ctx, func() error {
    return flushMetrics()
}, trigger)
w.Run()

// Run a typed callback for each value received on the channel.
jobs := make(chan Job)
w = workers.NewWorkerOnEvent(ctx, func(j Job) error {
    return process(j)
}, jobs)
w.Run()
```

Utility functions:

```go
// Retry a fallible operation up to n times, stopping early on ctx cancel.
err := workers.DoWithRetries(ctx, func() error {
    return connectToDB()
}, 3)

// Retry with explicit back-off delays between each attempt.
err = workers.DoWithDelays(ctx, func() error {
    return connectToDB()
}, 1*time.Second, 5*time.Second, 30*time.Second)

// Measure wall-clock execution time.
duration, err := workers.DoWithStopwatch(func() error {
    return runMigration()
})

// Sleep that is interrupted early when ctx is cancelled.
workers.Sleep(ctx, 5*time.Second)
```

---

### `collections` — generic data structures

All three implementations satisfy a common `Collection[I, T]` interface
(`Size`, `IsEmpty`, `At`, `Get`, `Push`, `PushAll`, `Join`, `Merge`,
`Delete`, `DeleteBy`, `DeleteAll`, `Some`, `Find`, `Filter`, `ForEach`).

#### `collections/array`

A slice-backed ordered collection. Pass an optional preSize to reserve
capacity upfront and avoid repeated re-allocations.

```go
a := array.New[int](100)       // pre-allocate capacity for 100 elements
a  = array.From(1, 2, 3)
a  = array.Wrap(existingSlice)

a.Push(4)
a.PushAll(5, 6, 7)
a.At(0)                        // raw index, no bounds check
a.Get(-1)                      // last element; negative indices count from end
a.First() / a.Last()
a.Slice(1, 4)                  // returns a new array [1:4)
a.Delete(i)                    // O(1) swap-with-last
a.DeleteKeepOrdering(i, true)  // O(n) copy-based, preserves order

// Sorting and searching (collections/array extension)
cmp := func(a, b int) int { return a - b }

a.InsertionSort(cmp)           // O(n²), good for small / nearly-sorted data
a.HeapSort(cmp)                // O(n log n), O(1) extra space
a.IsSorted(cmp)                // true if ascending or descending

v, ok := a.BinarySearch(42, cmp) // O(log n), array must be sorted
min, ok := a.Min(cmp)
max, ok := a.Max(cmp)
```

#### `collections/set`

A map-backed collection of unique elements. Elements must implement
`Keyable[I]`; primitive types can use the built-in wrappers.

```go
// Custom Keyable type
type UserID struct{ id string }
func (u *UserID) Key() string { return u.id }

s := set.New[string, *UserID](50) // optional preSize capacity hint
s  = set.From(item1, item2)

// Primitives via built-in wrapper
ps := set.NewPrimitiveSet[int]()
ps  = set.FromPrimitives(1, 2, 3)
ps  = set.WrapPrimitives([]int{1, 2, 3})

s.Push(item)           // no-op if the key already exists
s.Has(item)            // O(1) membership check
s.At(key) / s.Get(key) // retrieve by key
s.Delete(key)
s.Filter(predicate)    // returns a new set
s.Merge(other)         // returns a new set, originals unchanged
```

#### `collections/linkedlist`

A doubly-linked list with O(1) head/tail operations and bidirectional
traversal optimisation for index-based access.

```go
list := linkedlist.NewLinkedList[int]()

list.Push(1)        // append to tail  — O(1)
list.PushFront(0)   // insert at head  — O(1)
list.PushAll(2, 3, 4)

list.First() / list.Last()
list.At(0) / list.At(-1)   // negative indices count from the end
list.Pop(i)                // remove and return element at index i
list.PopLeft()             // remove head — O(1)
list.PopRight()            // remove tail — O(1)

list.Swap(i, j)            // handles adjacent and non-adjacent nodes
list.Move(from, to)
list.Shrink(n)             // trim to n elements from the front

list.Sort(cmp)             // in-place quicksort — O(n log n) average

// Cache-friendly node API — O(1) operations with a held reference
node := list.Insert(v)    // append, returns *LinkedNode
node  = list.InsertFront(v)
list.MoveToFront(node)
list.Remove(node)
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
