package housekeeper

import (
	"sync"
	"testing"
	"time"

	"github.com/elastic/beats/filebeat/input"
	"github.com/elastic/beats/filebeat/input/file"
	"github.com/stretchr/testify/assert"
)

type eventSource struct {
	event input.Event
	out   chan<- []*input.Event
	wg    sync.WaitGroup
	done  chan struct{}
}

func (s *eventSource) Start() {
	s.wg.Add(1)
	go s.Run()
}

func (s *eventSource) Run() {
	defer func() {
		s.wg.Done()
	}()
	for {
		select {
		case <-s.done:
			return
		default:
			s.event.State.TTL = 1 * time.Hour
			s.out <- []*input.Event{&s.event}
		}
	}
}

func (s *eventSource) Stop() {
	close(s.done)
	s.wg.Wait()
}

func Test_HouseKeeper(t *testing.T) {
	expected := "/var/log/serviceA/b.log"

	states := []file.State{
		// active file
		{
			Source:   "/var/log/serviceA/a.log",
			TTL:      1 * time.Hour,
			Finished: false,
		},
		// timeout inactive file, not last file in dir
		// expected to be deleted
		{
			Source:   expected,
			TTL:      0,
			Finished: true,
		},
		// timeout inactive file, last file in dir
		{
			Source:   "/var/log/serviceB/stdout",
			TTL:      0,
			Finished: true,
		},
	}
	oldStates := file.States{}
	oldStates.SetStates(states)

	ch := make(chan []*input.Event)
	// start eventSource
	eSource := eventSource{
		event: input.Event{
			InputType: "log",
			State: file.State{
				Source: "/var/log/serviceA/a.log",
				TTL:    1 * time.Hour,
			},
		},
		out:  ch,
		done: make(chan struct{}),
	}
	eSource.Start()

	actual := []string{}
	rmFunc := func(name string) error {
		actual = append(actual, name)
		return nil
	}
	hk := New(ch, &oldStates, rmFunc, 1)
	hk.Start()

	time.Sleep(5 * time.Second)

	eSource.Stop()
	hk.Stop()

	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expected, actual[0])
}
