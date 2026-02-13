package core

import (
	"sync"

	"github.com/rs/zerolog"
)

type Dispatcher struct {
	outputs []OutputPort
	logger  zerolog.Logger
}

func NewDispatcher(outputs []OutputPort, logger *zerolog.Logger) *Dispatcher {
	log := logger.
		With().
		Str("Componet", "ddispatcher").
		Logger()

	return &Dispatcher{
		outputs: outputs,
		logger:  log,
	}
}

func (d *Dispatcher) Dispatch(event Event) {
	d.logger.
		Debug().
		Str("event", event.Id).
		Str("type", event.Type).
		Str("repo", event.Repository.FullName).
		Int("outputs", len(d.outputs)).
		Msg("Dispatching event")

	var wg sync.WaitGroup
	var mu sync.Mutex

	failed := make([]string, 0)

	for _, output := range d.outputs {
		wg.Add(1)
		go func(out OutputPort) {
			defer wg.Done()

			if err := out.SendMessage(event); err != nil {
				d.logger.
					Error().
					Err(err).
					Str("output", out.Name()).
					Msg("Failed to send event to output")
				mu.Lock()
				failed = append(failed, out.Name())
				mu.Unlock()
				return
			}

			d.logger.
				Debug().
				Str("output", out.Name()).
				Str("event", event.Id).
				Msg("Event dispatched")
		}(output)
	}

	wg.Wait()

	d.logger.
		Info().
		Int("failed", len(failed)).
		Int("outputs", len(d.outputs)).
		Msg("Event dispatched to all outputs")
}

func (d *Dispatcher) AddOutput(port OutputPort) {
	d.outputs = append(d.outputs, port)
	d.logger.Info().Str("output", port.Name()).Msg("Output added")
}

func (d *Dispatcher) OutputNames() []string {
	names := make([]string, len(d.outputs))
	for i, output := range d.outputs {
		names[i] = output.Name()
	}

	return names
}
