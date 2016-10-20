package housekeeper

import (
	"sync"
	"time"

	input "github.com/elastic/beats/filebeat/input"

	cfg "github.com/elastic/beats/filebeat/config"
	"github.com/elastic/beats/filebeat/input/file"
	"github.com/elastic/beats/libbeat/logp"
)

type HouseKeeper struct {
	in              <-chan []*input.Event
	done            chan struct{}
	states          *file.States
	wg              sync.WaitGroup
	cleanupInterval int64
	remove          func(name string) error
}

func New(in chan []*input.Event, states *file.States, rmFunc func(name string) error, cleanupInterval int64) *HouseKeeper {
	return &HouseKeeper{
		in:              in,
		done:            make(chan struct{}, 10),
		states:          states,
		cleanupInterval: cleanupInterval,
		remove:          rmFunc,
	}
}

func (h *HouseKeeper) Start() {
	h.wg.Add(1)
	go h.Run()
}

func (h *HouseKeeper) Run() {
	logp.Info("Starting HouseKeeper")

	defer func() {
		h.wg.Done()
	}()

	last := time.Now().Unix()
	for {
		var events []*input.Event

		select {
		case <-h.done:
			logp.Info("Ending HouseKeeper")
			return
		case events = <-h.in:
		}

		for _, event := range events {
			if event.InputType == cfg.StdinInputType {
				continue
			}
			h.states.Update(event.State)
		}

		current := time.Now().Unix()
		if current-last >= h.cleanupInterval {
			h.Cleanup()
			last = current
		}
	}
}

func (h *HouseKeeper) Cleanup() {
	states := h.states.GetStates()

	// key: dirname, value: file states
	dirs := make(map[string]*DirFilesState)

	for _, state := range states {
		key := dirname(state.Source)
		if dirs[key] == nil {
			dirs[key] = &DirFilesState{states: []file.State{}}
		}
		dirs[key].Add(state)
	}

	timeoutStates := []file.State{}
	h.states.CleanupWithFunc(func(state file.State) {
		timeoutStates = append(timeoutStates, state)
	})

	// TODO: for stderr
	// TODO: for stdout
	// TODO: remove file size > xxx
	for _, state := range timeoutStates {
		dir := dirname(state.Source)
		if dirs[dir].Len() > 1 {
			err := h.remove(state.Source)
			if err != nil {
				logp.Err("remove file failed, err[%s]", err)
			}
			dirs[dir].Remove(state)
		} else {
			logp.Info("last inactive log file[%s], will not delete", state.Source)
		}
	}
}

func (h *HouseKeeper) Stop() {
	logp.Info("Stopping HouseKeeper")
	close(h.done)
	h.wg.Wait()
}
