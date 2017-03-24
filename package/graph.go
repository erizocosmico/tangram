package pkg

import "fmt"

// Graph represents a dependency graph.
type Graph struct {
	// Node is the root node of the graph, which is always one module.
	root  *node
	nodes map[string]*node
}

// NewGraph creates a new graph with the given root module.
func NewGraph(root string) *Graph {
	rootNode := newNode(root)
	return &Graph{
		root: rootNode,
		nodes: map[string]*node{
			root: rootNode,
		},
	}
}

// Add adds `to` as a dependency of `from`.
func (g *Graph) Add(to, from string) *Graph {
	g.node(from).add(g.node(to))
	return g
}

func (g *Graph) node(module string) *node {
	if n, ok := g.nodes[module]; ok {
		return n
	}

	n := newNode(module)
	g.nodes[module] = n
	return n
}

// Resolve returns a list of nodes in the exact order in which they need to be
// resolved. A graph with the exact same nodes in the exact same order produces
// an output exactly equal no matter how many times it's called.
func (g *Graph) Resolve() ([]string, error) {
	ctx := newResolutionCtx()
	if err := g.root.resolve(ctx); err != nil {
		return nil, err
	}

	return ctx.nodes, nil
}

type node struct {
	module     string
	edges      map[string]*node
	dependants []string
}

func newNode(module string) *node {
	return &node{
		module: module,
		edges:  make(map[string]*node),
	}
}

func (n *node) add(node *node) {
	if _, ok := n.edges[node.module]; !ok {
		n.edges[node.module] = node
		n.dependants = append(n.dependants, node.module)
	}
}

func (n *node) resolve(ctx *resolutionCtx) error {
	ctx.unresolved.add(n.module)

	for _, mod := range n.dependants {
		if !ctx.resolved.contains(mod) {
			if ctx.unresolved.contains(mod) {
				return NewCircularDependencyError(n.module, mod)
			}

			if err := n.edges[mod].resolve(ctx); err != nil {
				return err
			}
		}
	}

	delete(ctx.unresolved, n.module)
	ctx.resolved.add(n.module)
	ctx.nodes = append(ctx.nodes, n.module)
	return nil
}

type moduleSet map[string]struct{}

func (m moduleSet) add(module string) {
	m[module] = struct{}{}
}

func (m moduleSet) contains(module string) bool {
	_, ok := m[module]
	return ok
}

type resolutionCtx struct {
	nodes      []string
	unresolved moduleSet
	resolved   moduleSet
}

func newResolutionCtx() *resolutionCtx {
	return &resolutionCtx{
		unresolved: make(moduleSet),
		resolved:   make(moduleSet),
	}
}

// CircularDependencyError describes an error because there was a circular
// dependency between two dependencies.
type CircularDependencyError struct {
	// Modules that depended on each other.
	Modules [2]string
}

// NewCircularDependencyError returns a new CircularDependencyError.
func NewCircularDependencyError(a, b string) *CircularDependencyError {
	return &CircularDependencyError{[2]string{a, b}}
}

func (e CircularDependencyError) Error() string {
	return fmt.Sprintf(
		"circular dependency error: %s -> %s",
		e.Modules[0],
		e.Modules[1],
	)
}
