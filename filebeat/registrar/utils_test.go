package registrar

import (
	"testing"

	"github.com/elastic/beats/filebeat/input/file"
	"github.com/stretchr/testify/assert"
)

func Test_dirname(t *testing.T) {
	filenames := []string{
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID/stdout",
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID/stderr",
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID/log/0/a",
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID/log/1/b",
	}
	dirs := []string{
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID",
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID",
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID/log/0",
		"/disk1/mesos/slaves/SLAVEID/frameworks/FWID/executors/TASKID/runs/UUID/log/1",
	}
	for i, filename := range filenames {
		dir := dirname(filename)
		assert.Equal(t, dirs[i], dir)
	}
}

func Test_DirFilesState(t *testing.T) {
	state := &DirFilesState{states: []file.State{}}
	assert.Equal(t, 0, state.Len())

	state.Remove(file.State{Source: "a"})

	state.Add(file.State{Source: "a"})
	state.Add(file.State{Source: "b"})
	assert.Equal(t, 2, state.Len())

	state.Remove(file.State{Source: "a"})
	assert.Equal(t, 1, state.Len())
	state.Remove(file.State{Source: "c"})
	assert.Equal(t, 1, state.Len())
	state.Remove(file.State{Source: "b"})
	assert.Equal(t, 0, state.Len())
}
