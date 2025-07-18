package dice

type Component struct {
	// Name of the component
	Name string
	// An adapter to call operations on the cosmos database
	Adapter CosmosAdapter
	// A wraper for getting the relevant nodes
	// NOTE: graphs are also nodes, just wrapped (GraphNode)
	Nodes func(e Event) []GraphNode
	// Types of events the component listens to
	Events []EventType
}

// Sends the event to the modules to handle.
// If the event points to some object with hooks,
// the event is only pushed to the hookers
func (c *Component) update(e Event) error {
	for _, n := range c.Nodes(e) {
		if err := n.Update(e); err != nil {
			return err
		}
	}
	return nil
}
