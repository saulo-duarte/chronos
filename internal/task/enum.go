package task

type TaskStatus string
type TaskPriority string
type TaskType string

const (
	TODO        TaskStatus = "TODO"
	DONE        TaskStatus = "DONE"
	IN_PROGRESS TaskStatus = "IN_PROGRESS"

	EVENT   TaskType = "EVENT"
	PROJECT TaskType = "PROJECT"
	STUDY   TaskType = "STUDY"

	LOW    TaskPriority = "LOW"
	HIGH   TaskPriority = "HIGH"
	MEDIUM TaskPriority = "MEDIUM"
)
