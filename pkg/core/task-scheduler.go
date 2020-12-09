package core

import (
	"time"
)

type Task struct {
	Id           string
	ExecDelay    int64
	LastExec     int64 `default:"0"`
	DelayedStart bool  `default:"false"`
	Enabled      bool  `default:"true"`
	Executing    bool  `default:"false"`
	TaskFunc     func()
}

type TaskScheduler struct {
	tasks []Task
}

func (ts *TaskScheduler) StartServerTickLoop() {
	for {
		for c, task := range ts.tasks {
			if task.LastExec == 0 && task.DelayedStart { // Newly added task; Check for delayed start.
				ts.tasks[c].LastExec = (time.Now().UnixNano() / 1000000)
				continue
			}

			// Check if task needs to be executed
			if task.LastExec+task.ExecDelay < time.Now().UnixNano()/1000000 {
				if task.Executing {
					continue
				}

				ts.tasks[c].LastExec = time.Now().UnixNano() / 1000000

				ts.tasks[c].Executing = true
				task.TaskFunc()
				ts.tasks[c].Executing = false
			}
		}
		time.Sleep(50) // 20 ticks in a second; 1000 / 20
	}
}

func (s *TaskScheduler) AddTask(task Task) {
	s.tasks = append(s.tasks, task)
}

func (s *TaskScheduler) DisableTask(taskId string) bool {
	for _, task := range s.tasks {
		if task.Id == taskId {
			task.Enabled = false
			return true
		}
	}
	return false
}

func (s *TaskScheduler) EnableTask(taskId string) bool {
	for _, task := range s.tasks {
		if task.Id == taskId {
			task.Enabled = true
			return true
		}
	}
	return false
}
