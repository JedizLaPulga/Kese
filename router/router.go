package router

import (
	"strings"
)

// HandlerFunc is copied here to avoid circular imports.
// It must match the signature in the main kese package.
type HandlerFunc interface{}

type Router struct {
	trees map[string]*node // one tree per HTTP method
}

// node represents a node in the routing tree.
// It can represent a static path segment, a parameter, or a wildcard.
type node struct {
	// path is the path segment this node represents
	path string

	// children are the static child nodes
	children map[string]*node

	// paramChild is the child node for a parameter (e.g., :id)
	paramChild *node

	// paramName is the name of the parameter if this is a param node
	paramName string

	// handler is the handler function for this route (if this is a leaf node)
	handler HandlerFunc

	// isLeaf indicates if this node represents a complete route
	isLeaf bool
}

// New creates a new Router instance.
func New() *Router {
	return &Router{
		trees: make(map[string]*node),
	}
}

// Add registers a new route with the given method, path, and handler.
// Path can contain parameters in the format ":paramName" (e.g., "/users/:id").
func (r *Router) Add(method, path string, handler HandlerFunc) {
	// Get or create the tree for this HTTP method
	root, exists := r.trees[method]
	if !exists {
		root = &node{
			path:     "/",
			children: make(map[string]*node),
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
				current.paramChild = &node{
					path:      segment,
					paramName: paramName,
					children:  make(map[string]*node),
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
				child = &node{
					path:     segment,
					children: make(map[string]*node),
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
func (r *Router) Match(method, path string) (HandlerFunc, map[string]string) {
	// Get the tree for this HTTP method
	root, exists := r.trees[method]
	if !exists {
		return nil, nil
	}

	params := make(map[string]string)

	// Handle root path
	if path == "/" {
		if root.isLeaf {
			return root.handler, params
		}
		return nil, nil
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
			params[current.paramChild.paramName] = segment
			current = current.paramChild
			continue
		}

		// No match found
		return nil, nil
	}

	// Check if we're at a leaf node
	if current.isLeaf {
		return current.handler, params
	}

	return nil, nil
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
