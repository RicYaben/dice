package dice

import (
	"github.com/pkg/errors"
)

//const EngineID uint = 0xD1CE

type engine struct {
	adapter EngineAdapter
	emitter Emitter
}

func NewEngine(adapter EngineAdapter, emitter Emitter) *engine {
	return &engine{adapter, emitter}
}

// Entrypoint to start the engine.
func (e *engine) Run(sources []*Source) error {
	for _, src := range sources {
		if err := e.adapter.AddSource(src); err != nil {
			return errors.Wrapf(err, "failed to consume events from source %v", src)
		}
	}
	return nil
}
