package router

import (
	"strings"
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// Get returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) Get(name string) string {
	for _, entry := range ps {
		if entry.Key == name {
			return entry.Value
		}
	}
	return ""
}

// Router is a generic radix tree router.
type Router[T any] struct {
	trees map[string]*node[T] // one tree per HTTP method
}

// node represents a node in the routing tree.
// It can represent a static path segment, a parameter, or a wildcard.
type node[T any] struct {
	// path is the path segment this node represents
	path string

	// children are the static child nodes
	children map[string]*node[T]

	// paramChild is the child node for a parameter (e.g., :id)
	paramChild *node[T]

	// paramName is the name of the parameter if this is a param node
	paramName string

	// handler is the handler function for this route (if this is a leaf node)
	handler T

	// isLeaf indicates if this node represents a complete route
	isLeaf bool
}

// New creates a new Router instance.
func New[T any]() *Router[T] {
	return &Router[T]{
		trees: make(map[string]*node[T]),
	}
}

// Add registers a new route with the given method, path, and handler.
// Path can contain parameters in the format ":paramName" (e.g., "/users/:id").
func (r *Router[T]) Add(method, path string, handler T) {
	// Get or create the tree for this HTTP method
	root, exists := r.trees[method]
	if !exists {
		root = &node[T]{
			path:     "/",
			children: make(map[string]*node[T]),
		}
		r.trees[method] = root
	}

	// If path is just "/", register at root
	if path == "/" {
		root.handler = handler
		root.isLeaf = true
		return
	}

	// Split path into segments
	segments := splitPath(path)
	current := root

	// Traverse/build the tree
	for i, segment := range segments {
		isLast := i == len(segments)-1

		// Check if this is a parameter segment
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:] // remove the ":"

			// Create or get param child
			if current.paramChild == nil {
				current.paramChild = &node[T]{
					path:      segment,
					paramName: paramName,
					children:  make(map[string]*node[T]),
				}
			}

			current = current.paramChild

			if isLast {
				current.handler = handler
				current.isLeaf = true
			}
		} else {
			// Static segment
			child, exists := current.children[segment]
			if !exists {
				child = &node[T]{
					path:     segment,
					children: make(map[string]*node[T]),
				}
				current.children[segment] = child
			}

			current = child

			if isLast {
				current.handler = handler
				current.isLeaf = true
			}
		}
	}
}

// Match finds a handler that matches the given method and path.
// It returns the handler and any extracted parameters.
// The third return value indicates whether a match was found.
func (r *Router[T]) Match(method, path string) (T, Params, bool) {
	var zero T
	// Get the tree for this HTTP method
	root, exists := r.trees[method]
	if !exists {
		return zero, nil, false
	}

	params := make(Params, 0) // TODO: Optimize this with a sync.Pool if needed later

	// Handle root path
	if path == "/" {
		if root.isLeaf {
			return root.handler, params, true
		}
		return zero, nil, false
	}

	// Split path into segments
	segments := splitPath(path)
	current := root

	// Traverse the tree
	for _, segment := range segments {
		// Try static match first
		if child, exists := current.children[segment]; exists {
			current = child
			continue
		}

		// Try parameter match
		if current.paramChild != nil {
			// Check if isLeaf to ensure we don't match partial paths if we don't have to?
			// Actually standard logic is greedy. match param.
			params = append(params, Param{Key: current.paramChild.paramName, Value: segment})
			current = current.paramChild
			continue
		}

		// No match found
		return zero, nil, false
	}

	// Check if we're at a leaf node
	if current.isLeaf {
		return current.handler, params, true
	}

	return zero, nil, false
}

// splitPath splits a path into segments, removing empty segments.
// For example: "/users/:id/posts" -> ["users", ":id", "posts"]
func splitPath(path string) []string {
	// Remove leading and trailing slashes
	path = strings.Trim(path, "/")

	if path == "" {
		return []string{}
	}

	segments := strings.Split(path, "/")

	// Filter out empty segments
	result := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment != "" {
			result = append(result, segment)
		}
	}

	return result
}
