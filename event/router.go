package event

import "errors"

// Router dispatches events to registered handlers by event ID.
// It supports both synchronous and asynchronous dispatch.
type Router struct {
	async   bool
	routes  map[string]Handler
	onError ErrorHandler
}

// RouterParams configures a Router.
type RouterParams struct {
	// Async enables asynchronous handler dispatch. When true, each handler is
	// invoked in its own goroutine.
	Async bool
	// OnErr is called whenever a handler returns an error or no handler is
	// registered for a dispatched event. May be nil.
	OnErr ErrorHandler
}

// NewRouter creates a new Router. An optional RouterParams may be provided to
// configure async dispatch and error handling; zero values are used otherwise.
func NewRouter(params ...RouterParams) *Router {
	var p RouterParams
	if len(params) > 0 {
		p = params[0]
	}

	return &Router{
		async:   p.Async,
		routes:  make(map[string]Handler),
		onError: p.OnErr,
	}
}

// OnEvent registers handler for events with the given id. Calling OnEvent with
// the same id more than once replaces the previous handler.
func (router *Router) OnEvent(id ID, handler Handler) {
	router.routes[id] = handler
}

// Route dispatches event to its registered handler. If no handler is registered
// and an ErrorHandler was configured, it is called with a "no handler registered"
// error. When async mode is enabled the handler runs in a new goroutine.
func (router *Router) Route(event Event) {
	if handler, ok := router.routes[event.Id()]; ok {
		if router.async {
			go router.call(event, handler)
			return
		}
		router.call(event, handler)
	} else {
		if router.onError != nil {
			router.onError(event, errors.New("no handler registered"))
		}
	}
}

func (router *Router) call(event Event, handler Handler) {
	err := handler(event.Data())
	if err != nil && router.onError != nil {
		router.onError(event, err)
	}
}
