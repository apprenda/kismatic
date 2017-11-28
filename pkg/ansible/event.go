package ansible

// Event produced by Ansible when running a playbook
type Event interface {
	// Type is the name of the event type
	Type() string
}

type namedEvent struct {
	Name string
}

type runnerResult struct {
	// Command is the command that was run
	Command []string `json:"cmd"`
	// Stdout captured when the command was run
	Stdout string
	// Stderr captured when the command was run
	Stderr string
	// Message returned by the runner
	Message string `json:"msg"`
	// Item that corresponds to this result. Available only when event is related
	// to an item
	Item string
	// Number of attempts a task has been retried
	Attempts int
	// Maximum number of retries for a given task
	MaxRetries int `json:"retries"`
}

type runnerResultEvent struct {
	Host         string
	Result       runnerResult
	IgnoreErrors bool
}

// PlaybookStartEvent signals the beginning of a playbook
type PlaybookStartEvent struct {
	namedEvent
	Count int
}

func (e *PlaybookStartEvent) Type() string {
	return "Playbook Start"
}

// PlaybookEndEvent signals the beginning of a playbook
type PlaybookEndEvent struct {
	namedEvent
}

func (e *PlaybookEndEvent) Type() string {
	return "Playbook End"
}

// PlayStartEvent signals the beginning of a play
type PlayStartEvent struct {
	namedEvent
}

func (e *PlayStartEvent) Type() string {
	return "Play Start"
}

// TaskStartEvent signals the beginning of a task
type TaskStartEvent struct {
	namedEvent
}

func (e *TaskStartEvent) Type() string {
	return "Task Start"
}

// HandlerTaskStartEvent signals the beginning of a handler task
type HandlerTaskStartEvent struct {
	namedEvent
}

func (e *HandlerTaskStartEvent) Type() string {
	return "Handler Task Start"
}

// RunnerOKEvent signals the successful completion of a runner
type RunnerOKEvent struct {
	runnerResultEvent
}

func (e *RunnerOKEvent) Type() string {
	return "Runner OK"
}

// RunnerFailedEvent signals a failure when executing a runner
type RunnerFailedEvent struct {
	runnerResultEvent
}

func (e *RunnerFailedEvent) Type() string {
	return "Runner Failed"
}

// RunnerItemOKEvent signals the successful completion of a runner item
type RunnerItemOKEvent struct {
	runnerResultEvent
}

func (e *RunnerItemOKEvent) Type() string {
	return "Runner Item OK"
}

// RunnerItemFailedEvent signals the failure of a task with a specific item
type RunnerItemFailedEvent struct {
	runnerResultEvent
}

func (e *RunnerItemFailedEvent) Type() string {
	return "Runner Item Failed"
}

// RunnerItemRetryEvent signals the retry of a runner item
type RunnerItemRetryEvent struct {
	runnerResultEvent
}

func (e *RunnerItemRetryEvent) Type() string {
	return "Runner Item Retry"
}

// RunnerSkippedEvent is raised when a runner is skipped
type RunnerSkippedEvent struct {
	runnerResultEvent
}

func (e *RunnerSkippedEvent) Type() string {
	return "Runner Skipped"
}

// RunnerUnreachableEvent is raised when the target host is not reachable via SSH
type RunnerUnreachableEvent struct {
	runnerResultEvent
}

func (e *RunnerUnreachableEvent) Type() string {
	return "Runner Unreachable"
}
