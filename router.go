package cxr

import "fmt"

// Router resolves a Request to a Handler.
//
// Resolution chain (first match wins):
//  1. Explicit handler ID — Request.Config["handler"].(string)
//  2. Capability intersection — any m ∈ handler.caps ∩ req.Capabilities
//  3. Tools intersection     — any m ∈ handler.tools ∩ req.Tools
//  4. Default handler        — first registered handler
type Router struct {
	handlers []Handler
}

// NewRouter returns a Router backed by handlers (order matters for default).
func NewRouter(handlers []Handler) *Router {
	return &Router{handlers: handlers}
}

// Route returns the first Handler that can service req, or an error if none.
func (r *Router) Route(req Request) (Handler, error) {
	// 1. explicit ID override
	if id, ok := req.Config["handler"].(string); ok && id != "" {
		for _, h := range r.handlers {
			if h.ID() == id {
				return h, nil
			}
		}
		return nil, fmt.Errorf("cxr: no handler with id %q registered", id)
	}

	// 2+3. capability/tool intersection via CanHandle
	for _, h := range r.handlers {
		if h.CanHandle(req) {
			return h, nil
		}
	}

	// 4. default
	if len(r.handlers) > 0 {
		return r.handlers[0], nil
	}

	return nil, fmt.Errorf("cxr: no handlers registered")
}
