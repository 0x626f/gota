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

#### `collections/graph`

Three interchangeable topology backends and a typed `Graph[Vertex, Edge, Key]`
wrapper. All share a single interface; swap the backend via `TopologyParams`.

**Topology** — low-level, key-only API

```go
// Pick a backend; all three satisfy ITopology[Key].
t := graph.NewAdjacencyMatrix[string]() // map-of-maps, O(1) lookup
t  = graph.NewBitMatrix[string]()       // compact bit-rows, cache-friendly
t  = graph.NewCSR[string]()             // sorted adjacency lists, compact memory

// Via NewTopology — selects backend from params.
t = graph.NewTopology[string](graph.TopologyParams{
    Key:      graph.CRSTopology,
    Features: graph.Features(graph.Directed),
    Scale:    256, // pre-allocates capacity for ~256 vertices
})

t.Add("A")
t.Set("A", "B")          // creates both vertices implicitly
t.Has("A", "B")          // true — edge exists
t.Has("A", "B", "C")     // true — path A→B→C exists
t.Remove("A", "B", "C")  // remove edges along the path
t.Delete("A")            // remove vertex and all incident edges
t.Neighbors("B")         // []string{"C"}
t.IsCycled()             // false
```

**Graph** — typed wrapper with vertex/edge data and traversal

```go
// Implement IVertex for your domain type.
type City struct{ name string }
func (c City) Key() string { return c.name }

// Edge type can be anything — use struct{} if you only need topology.
type Road struct{ km float64 }

// Key (string) is inferred from City; only Vertex and Edge need to be explicit.
g := graph.NewGraph[City, Road](graph.TopologyParams{
    Key:      graph.AdjacencyMatrixTopology,
    Features: graph.Features(graph.Directed),
})

london, paris, berlin := City{"London"}, City{"Paris"}, City{"Berlin"}
g.Set(london, paris,  Road{340})
g.Set(paris,  berlin, Road{878})

v, _ := g.GetVertex("London")    // IVertex[string]
e, _ := g.GetEdge(london, paris) // Road{340}
g.Neighbors(london)               // []string{"Paris"}

// Traversal — embed BlankTopologyTraversal and override only what you need.
type visitor struct{ graph.BlankTopologyTraversal[string] }
func (v *visitor) ShouldVisit(_, _ string) bool         { return true }
func (v *visitor) OnVisit(from, to string) bool         { fmt.Println(from, "→", to); return true }
func (v *visitor) EdgeWeight(_, _ string) graph.Weight  { return 1 }

g.DFS("London", &visitor{})
g.BFS("London", &visitor{})
g.Dijkstra("London", &visitor{})
g.AStar("London", "Berlin", &visitor{})

// All simple paths from London.
paths := g.Paths("London", false, graph.DFSSearch)
// paths == [["London" "Paris" "Berlin"]]
```

**Benchmarks** — `goos: linux`, `goarch: amd64`, `cpu: 12th Gen Intel(R) Core(TM) i5-12500H`

Δ = `(AM − X) / AM × 100` relative to AdjacencyMatrix — positive means faster.
B/op shown as `AM / BM / CSR` when values differ across implementations.
† Delete and Remove benchmarks measure a remove+restore cycle (delete+add / remove+set / remove+rebuild).

| Operation           | AM ns/op | BM ns/op |       BM Δ | CSR ns/op |      CSR Δ | B/op          | allocs      |
|---------------------|---------:|---------:|-----------:|----------:|-----------:|--------------:|------------:|
| Add                 |    205.3 |    64.53 | **+68.6 %** |     68.56 | **+66.6 %** | 48 / 0 / 0  | 1 / 0 / 0   |
| Add (existing)      |    2.795 |    3.341 |    −19.5 % |     3.284 |    −17.5 % |             0 |           0 |
| Contains (present)  |    2.524 |    2.549 |     −1.0 % |     2.503 |     +0.8 % |             0 |           0 |
| Contains (absent)   |    1.990 |    2.226 |    −11.9 % |     1.996 |     −0.3 % |             0 |           0 |
| Delete            |    121.0 |    54.16 | **+55.2 %** |     53.79 | **+55.5 %** | 48 / 0 / 0  | 1 / 0 / 0   |
| Set                 |    23.66 |    18.37 | **+22.4 %** |     22.93 |     +3.1 % |             0 |           0 |
| Set (directed)      |    22.35 |    23.98 |     −7.3 % |     26.38 |    −18.0 % |             0 |           0 |
| Has (vertex)        |    14.02 |    12.95 |     +7.6 % |     13.50 |     +3.7 % |             8 |           1 |
| Has (edge)          |    27.20 |    21.74 | **+20.1 %** |     21.71 | **+20.2 %** |            16 |           1 |
| Has path n=2        |    10.81 |    7.589 | **+29.8 %** |     7.525 | **+30.4 %** |             0 |           0 |
| Has path n=4        |    24.92 |    15.54 | **+37.6 %** |     14.88 | **+40.3 %** |             0 |           0 |
| Has path n=8        |    59.43 |    35.28 | **+40.6 %** |     30.88 | **+48.0 %** |             0 |           0 |
| Remove (vertex)   |    139.5 |    60.92 | **+56.3 %** |     58.56 | **+58.0 %** | 56 / 8 / 8  | 2 / 1 / 1   |
| Remove (edge)     |    84.47 |    53.12 | **+37.1 %** |     72.98 | **+13.6 %** |            16 |           1 |
| Remove path n=2   |    69.61 |    33.06 | **+52.5 %** |     49.25 | **+29.2 %** |             0 |           0 |
| Remove path n=4   |    194.6 |    93.55 | **+51.9 %** |     141.1 | **+27.4 %** |             0 |           0 |
| Remove path n=8   |    483.7 |    249.0 | **+48.5 %** |     343.5 | **+29.0 %** |             0 |           0 |

**`IsCycled`** — DFS-based cycle detection (B/op and allocs are identical across implementations at equal n)

| Variant | n | AM ns/op | BM ns/op | BM Δ | CSR ns/op | CSR Δ | B/op | allocs |
|---------|--:|---------:|---------:|-----:|----------:|------:|-----:|-------:|
| Undirected, no cycle | 8 | 896.0 | 514.3 | **+42.6 %** | 474.2 | **+47.1 %** | 176 | 9 |
| ↳ +presize | 8 | 901.2 | 558.6 | **+38.0 %** | 473.8 | **+47.4 %** | 176 | 9 |
| | 64 | 9076 | 8227 | +9.4 % | 5884 | **+35.2 %** | 6208 | 71 |
| ↳ +presize | 64 | 8954 | 8457 | +5.5 % | 6057 | **+32.4 %** | 6208 | 71 |
| | 512 | 73703 | 157831 | −114.1 % | 48787 | **+33.8 %** | 49216 | 519 |
| ↳ +presize | 512 | 72338 | 156489 | −116.3 % | 49742 | **+31.2 %** | 49216 | 519 |
| Undirected, cyclic | 8 | 984.6 | 575.2 | **+41.6 %** | 523.4 | **+46.8 %** | 336 | 11 |
| ↳ +presize | 8 | 975.0 | 574.6 | **+41.1 %** | 532.9 | **+45.3 %** | 336 | 11 |
| | 64 | 9398 | 8804 | +6.3 % | 6629 | **+29.5 %** | 7760 | 76 |
| ↳ +presize | 64 | 9239 | 8542 | +7.5 % | 6521 | **+29.4 %** | 7760 | 76 |
| | 512 | 78829 | 161680 | −105.1 % | 55057 | **+30.2 %** | 62224 | 527 |
| ↳ +presize | 512 | 77302 | 163353 | −111.3 % | 53168 | **+31.2 %** | 62224 | 527 |
| Directed DAG (no cycle) | 8 | 1069 | 812.9 | **+23.9 %** | 757.4 | **+29.1 %** | 559 | 14 |
| ↳ +presize | 8 | 1107 | 821.1 | **+25.8 %** | 773.3 | **+30.1 %** | 560 | 14 |
| | 64 | 7663 | 7822 | −2.1 % | 5350 | **+30.2 %** | ~4318 | 73 |
| ↳ +presize | 64 | 7656 | 7786 | −1.7 % | 5395 | **+29.5 %** | ~4309 | 73 |
| | 512 | 63279 | 156508 | −147.3 % | 44008 | **+30.4 %** | ~34300 | 524 |
| ↳ +presize | 512 | 63228 | 154834 | −144.9 % | 43706 | **+30.9 %** | ~34356 | 524 |
| Directed, cyclic | 8 | 927.1 | 548.0 | **+40.9 %** | 508.5 | **+45.1 %** | 328 | 14 |
| ↳ +presize | 8 | 939.3 | 556.1 | **+40.8 %** | 500.1 | **+46.7 %** | 328 | 14 |
| | 64 | 6791 | 7213 | −6.2 % | 4428 | **+34.8 %** | 4960 | 76 |
| ↳ +presize | 64 | 6790 | 7140 | −5.2 % | 4585 | **+32.5 %** | 4960 | 76 |
| | 512 | 56393 | 150124 | −166.2 % | 35300 | **+37.4 %** | 39712 | 527 |
| ↳ +presize | 512 | 54065 | 150682 | −178.7 % | 37568 | **+30.5 %** | 39712 | 527 |

**Acyclic `Set`** — cycle-enforcement overhead (union-find undirected, DFS reachability directed)

`+presize` rows construct the graph with a vertex-count hint to pre-allocate storage.

| Operation | n | AM ns/op | BM ns/op | BM Δ | CSR ns/op | CSR Δ | B/op | allocs |
|-----------|--:|---------:|---------:|-----:|----------:|------:|-----:|-------:|
| Set (acyclic undirected) | 8 | 1412 | 1211 | **+14.2 %** | 1050 | **+25.6 %** | 16 / 112 / 16 | 1 / 8 / 1 |
| ↳ +presize | 8 | 1375 | 1202 | **+12.6 %** | 1056 | **+23.2 %** | 16 / 112 / 16 | 1 / 8 / 1 |
| | 64 | 11323 | 12495 | −10.3 % | 8355 | **+26.2 %** | 16 / 1008 / 16 | 1 / 64 / 1 |
| ↳ +presize | 64 | 10866 | 12641 | −16.3 % | 8295 | **+23.7 %** | 16 / 1008 / 16 | 1 / 64 / 1 |
| | 512 | 91363 | 187892 | −105.7 % | 67270 | **+26.4 %** | 16 / 8176 / 16 | 1 / 512 / 1 |
| ↳ +presize | 512 | 90688 | 184977 | −103.9 % | 67015 | **+26.1 %** | 16 / 8176 / 16 | 1 / 512 / 1 |
| Set (acyclic directed) | 8 | 92.59 | 86.44 | +6.6 % | 96.31 | −4.0 % | 16 | 1 |
| ↳ +presize | 8 | 93.80 | 86.02 | +8.3 % | 93.49 | +0.3 % | 16 | 1 |
| | 64 | 109.4 | 95.60 | **+12.6 %** | 109.6 | −0.2 % | 16 | 1 |
| ↳ +presize | 64 | 106.6 | 95.50 | **+10.4 %** | 108.5 | −1.8 % | 16 | 1 |
| | 512 | 109.0 | 96.26 | **+11.7 %** | 103.8 | +4.8 % | 16 | 1 |
| ↳ +presize | 512 | 106.1 | 99.85 | +5.9 % | 104.5 | +1.5 % | 16 | 1 |
| Set reject (acyclic undirected) | 8 | 39.58 | 41.30 | −4.3 % | 41.74 | −5.5 % | 0 | 0 |
| ↳ +presize | 8 | 39.96 | 40.79 | −2.1 % | 40.36 | −1.0 % | 0 | 0 |
| | 64 | 44.28 | 45.16 | −2.0 % | 45.65 | −3.1 % | 0 | 0 |
| ↳ +presize | 64 | 45.12 | 47.11 | −4.4 % | 46.00 | −2.0 % | 0 | 0 |
| | 512 | 44.25 | 44.76 | −1.2 % | 45.29 | −2.4 % | 0 | 0 |
| ↳ +presize | 512 | 45.02 | 45.82 | −1.8 % | 45.17 | −0.3 % | 0 | 0 |
| Set reject (acyclic directed) | 8 | 345.4 | 227.1 | **+34.3 %** | 117.8 | **+65.9 %** | 0 / 56 / 0 | 0 / 7 / 0 |
| ↳ +presize | 8 | 347.3 | 225.5 | **+35.1 %** | 118.0 | **+66.0 %** | 0 / 56 / 0 | 0 / 7 / 0 |
| | 64 | 6481 | 7223 | −11.4 % | 4290 | **+33.8 %** | 4456 / 4960 / 4456 | 9 / 72 / 9 |
| ↳ +presize | 64 | 6609 | 7280 | −10.2 % | 4231 | **+36.0 %** | 4456 / 4960 / 4456 | 9 / 72 / 9 |
| | 512 | 51817 | 147422 | −184.5 % | 34230 | **+33.9 %** | 37320 / 41408 / 37320 | 15 / 526 / 15 |
| ↳ +presize | 512 | 52179 | 147905 | −183.5 % | 34918 | **+33.1 %** | 37320 / 41408 / 37320 | 15 / 526 / 15 |

**`Graph` struct** — higher-level wrapper over topology; includes vertex/edge maps and delegates traversal

| Operation        | AM ns/op | BM ns/op |       BM Δ | CSR ns/op |      CSR Δ | B/op          | allocs    |
|------------------|---------:|---------:|-----------:|----------:|-----------:|--------------:|----------:|
| Add              |    226.1 |    88.83 | **+60.7 %** |     87.54 | **+61.3 %** | 48 / 0 / 0  | 1 / 0 / 0 |
| Add (existing)   |    10.47 |    11.05 |     −5.5 % |     11.11 |     −6.1 % |             0 |         0 |
| Contains (pres.) |    4.498 |    4.710 |     −4.7 % |     4.566 |     −1.5 % |             0 |         0 |
| Contains (abs.)  |    4.759 |    4.763 |     −0.1 % |     4.710 |     +1.0 % |             0 |         0 |
| Delete         |    162.2 |    81.17 | **+49.9 %** |     77.80 | **+52.0 %** | 48 / 0 / 0  | 1 / 0 / 0 |
| Set              |    40.31 |    40.30 |      0.0 % |     42.90 |     −6.4 % |             0 |         0 |
| Has (vertex)     |    27.45 |    25.37 |     +7.6 % |     25.87 |     +5.8 % |            16 |         2 |
| Has (edge)       |    42.98 |    38.98 |     +9.3 % |     39.58 |     +7.9 % |            32 |         2 |
| GetVertex        |    5.216 |    5.299 |     −1.6 % |     5.245 |     −0.6 % |             0 |         0 |
| GetEdge          |    11.74 |    11.97 |     −2.0 % |     12.06 |     −2.7 % |             0 |         0 |
| Neighbors        |    51.32 |    22.71 | **+55.8 %** |     14.47 | **+71.8 %** |             8 |         1 |
| DFS              |   1067   |    877.8 | **+17.7 %** |     730.6 | **+31.5 %** |           504 |        14 |
| BFS              |    629.5 |    411.6 | **+34.6 %** |     352.6 | **+44.0 %** |           136 |        12 |
| Dijkstra         |   1183   |    965.3 | **+18.4 %** |     890.9 | **+24.7 %** |           336 |        24 |
| AStar            |   1195   |    974.4 | **+18.5 %** |     893.2 | **+25.2 %** |           336 |        24 |

**`Paths`** — all simple paths (DFS backtracking / BFS with copied state); chain n = path length

| Variant | n | AM ns/op | BM ns/op | BM Δ | CSR ns/op | CSR Δ | B/op | allocs |
|---------|--:|---------:|---------:|-----:|----------:|------:|-----:|-------:|
| DFS linear chain | 8 | 761.7 | 505.7 | **+33.6 %** | 475.7 | **+37.5 %** | 256 | 12 |
| | 32 | 4948 | 4035 | **+18.5 %** | 3520 | **+28.9 %** | 3144 | 45 |
| | 128 | 20049 | 24022 | −19.8 % | 14183 | **+29.3 %** | 13448 | 147 |
| BFS linear chain | 8 | 2087 | 1840 | **+11.8 %** | 1798 | **+13.8 %** | 2112 | 38 |
| | 32 | 20056 | 19875 | +0.9 % | 18763 | +6.4 % | 23616 | 206 |
| | 128 | 236397 | 244537 | −3.4 % | 234953 | +0.6 % | 335426 | 878 |
| DFS binary tree | 8 | 709.6 | 587.4 | **+17.2 %** | 560.4 | **+21.0 %** | 376 | 13 |
| | 32 | 2926 | 2441 | **+16.6 %** | 2223 | **+24.0 %** | 1872 | 40 |
| | 128 | 11796 | 12556 | −6.4 % | 8228 | **+30.2 %** | 8656 | 138 |
| BFS binary tree | 8 | 1839 | 1655 | **+10.0 %** | 1676 | +8.9 % | 2376 | 35 |
| | 32 | 8619 | 8230 | +4.5 % | 7594 | **+11.9 %** | 11096 | 125 |
| | 128 | 37283 | 38471 | −3.2 % | 33419 | **+10.4 %** | 45928 | 466 |
| DFS dense (cyclic) | 5 | 1406 | 1104 | **+21.5 %** | 1049 | **+25.4 %** | 824 | 23 |
| | 7 | 5619 | 4036 | **+28.2 %** | 3882 | **+30.9 %** | 3504 | 73 |
| | 9 | 25501 | 20401 | **+20.0 %** | 18999 | **+25.5 %** | 16264 | 271 |
| BFS dense (cyclic) | 5 | 4176 | 3647 | **+12.7 %** | 3560 | **+14.8 %** | ~5000 | 65 |
| | 7 | 17309 | 14965 | **+13.5 %** | 15054 | **+13.0 %** | ~21000 | 237 |
| | 9 | 75843 | 69788 | +8.0 % | 65514 | **+13.6 %** | ~86400 | 915 |
| DFS directed cycle | 8 | 849.4 | 587.5 | **+30.8 %** | 545.4 | **+35.8 %** | 392 | 15 |
| | 32 | 4918 | 4059 | **+17.5 %** | 3626 | **+26.3 %** | 3488 | 48 |
| | 128 | 20021 | 22724 | −13.5 % | 14405 | **+28.1 %** | 14656 | 150 |
| BFS directed cycle | 8 | 2296 | 2013 | **+12.3 %** | 1970 | **+14.2 %** | 2248 | 41 |
| | 32 | 20362 | 21051 | −3.4 % | 20968 | −3.0 % | 23960 | 209 |
| | 128 | 264318 | 255233 | +3.4 % | 262637 | +0.6 % | 336635 | 881 |

**`Graph` `Paths`** — same algorithm through the `Graph[V,E,K]` wrapper (n = number of vertices)

| Variant | n | AM ns/op | BM ns/op | BM Δ | CSR ns/op | CSR Δ | B/op | allocs |
|---------|--:|---------:|---------:|-----:|----------:|------:|-----:|-------:|
| DFS linear chain | 8 | 810.6 | 525.0 | **+35.2 %** | 475.9 | **+41.3 %** | 256 | 12 |
| | 32 | 4954 | 4038 | **+18.5 %** | 3535 | **+28.6 %** | 3144 | 45 |
| | 64 | 9855 | 9682 | +1.8 % | 7412 | **+24.8 %** | 6504 | 80 |
| BFS linear chain | 8 | 2089 | 1844 | **+11.7 %** | 1979 | +5.3 % | 2112 | 38 |
| | 32 | 22289 | 19353 | **+13.2 %** | 18824 | **+15.5 %** | 23616 | 206 |
| | 64 | 65500 | 65534 | −0.1 % | 63836 | +2.5 % | 86592 | 430 |

**DFS / BFS traversal** — topology-level callbacks; n = vertex count

| Variant | n | AM ns/op | BM ns/op | BM Δ | CSR ns/op | CSR Δ | B/op | allocs |
|---------|--:|---------:|---------:|-----:|----------:|------:|-----:|-------:|
| DFS linear chain | 64 | 12860 | 13029 | −1.3 % | 10845 | **+15.7 %** | 10432 | 88 |
| | 512 | 110893 | 215144 | −94.0 % | 84579 | **+23.7 %** | 86912 | 551 |
| | 4096 | 1087220 | 7067798 | −550.2 % | 861722 | **+20.7 %** | ~752870 | 4203 |
| DFS tree | 64 | 11167 | 10677 | +4.4 % | 9615 | **+13.9 %** | 9536 | 54 |
| | 512 | 87617 | 134469 | −53.5 % | 78307 | **+10.6 %** | 78976 | 291 |
| | 4096 | 827064 | 3756914 | −354.3 % | 714257 | **+13.6 %** | ~624866 | 2145 |
| DFS dense | 32 | 55651 | 35813 | **+35.7 %** | 35375 | **+36.4 %** | 65592 | 517 |
| | 128 | 1791630 | 1428562 | **+20.3 %** | 1344989 | **+24.9 %** | ~3203890 | 8159 |
| | 512 | 53243367 | 74928578 | −40.7 % | 56924653 | −6.9 % | ~194578000 | 130859 |
| BFS linear chain | 64 | 12152 | 12487 | −2.8 % | 9833 | **+19.1 %** | 10392 | 142 |
| | 512 | 96617 | 192108 | −98.8 % | 79306 | **+17.9 %** | 86872 | 1050 |
| | 4096 | 896260 | 6781592 | −656.7 % | 687186 | **+23.3 %** | ~690100 | 8280 |
| BFS tree | 64 | 10695 | 9863 | +7.8 % | 9316 | **+12.9 %** | 11368 | 58 |
| | 512 | 81725 | 123075 | −50.6 % | 71244 | **+12.8 %** | 102568 | 300 |
| | 4096 | 720148 | 3601782 | −400.1 % | 592580 | **+17.7 %** | ~814863 | 2161 |
| BFS dense | 32 | 49743 | 40744 | **+18.1 %** | 40200 | **+19.2 %** | 77680 | 986 |
| | 128 | 983334 | 747407 | **+24.0 %** | 722270 | **+26.5 %** | ~1615354 | 16171 |
| | 512 | 15650255 | 19826063 | −26.7 % | 17981935 | −14.9 % | ~28940400 | 261183 |

**Dijkstra / A\*** — lazy-deletion min-heap; n = vertex count

| Variant | n | AM ns/op | BM ns/op | BM Δ | CSR ns/op | CSR Δ | B/op | allocs |
|---------|--:|---------:|---------:|-----:|----------:|------:|-----:|-------:|
| Dijkstra linear chain | 64 | 23991 | 25370 | −5.7 % | 22668 | +5.5 % | 20400 | 228 |
| | 512 | 199148 | 282960 | −42.1 % | 173771 | **+12.7 %** | 169777 | 1596 |
| | 4096 | 1787216 | 7714635 | −331.7 % | 1495326 | **+16.3 %** | ~1347529 | 12472 |
| Dijkstra tree | 64 | 23126 | 23176 | −0.2 % | 22101 | +4.4 % | 21392 | 202 |
| | 512 | 192732 | 239114 | −24.1 % | 178185 | +7.5 % | 177937 | 1349 |
| | 4096 | 1803516 | 4900228 | −171.7 % | 1585598 | **+12.1 %** | ~1467611 | 10438 |
| Dijkstra dense | 32 | 40320 | 25421 | **+37.0 %** | 24372 | **+39.6 %** | 18712 | 130 |
| | 128 | 537317 | 289246 | **+46.2 %** | 289921 | **+46.0 %** | ~176664 | 436 |
| | 512 | 8555696 | 4157566 | **+51.4 %** | 4662456 | **+45.5 %** | ~2279200 | 1606 |
| A* linear chain | 64 | 24084 | 24823 | −3.1 % | 21476 | **+10.8 %** | 20400 | 228 |
| | 512 | 193441 | 282677 | −46.2 % | 177410 | +8.3 % | 169777 | 1596 |
| | 4096 | 1806387 | 7995672 | −342.5 % | 1536323 | **+14.9 %** | ~1347595 | 12472 |
| A* tree | 64 | 23568 | 22937 | +2.7 % | 21916 | +7.0 % | 21392 | 202 |
| | 512 | 194767 | 235118 | −20.7 % | 181641 | +6.7 % | 177937 | 1349 |
| | 4096 | 1820699 | 4882651 | −168.2 % | 1592049 | **+12.6 %** | ~1467612 | 10438 |
| A* dense | 32 | 24142 | 7314 | **+69.7 %** | 7080 | **+70.7 %** | 12532 / 8448 / 8448 | 92 / 64 / 64 |
| | 128 | 282749 | 29512 | **+89.6 %** | 32414 | **+88.5 %** | 99678 / 36288 / 36288 | 296 / 174 / 174 |
| | 512 | 4366269 | 132509 | **+96.9 %** | 140100 | **+96.8 %** | 1151360 / 144768 / 144768 | 1065 / 572 / 572 |

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
