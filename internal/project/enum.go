package project

type ProjectStatus string

const (
	NOT_INITIALIZED ProjectStatus = "NOT_INITIALIZED"
	IN_PROGRESS     ProjectStatus = "IN_PROGRESS"
	COMPLETED       ProjectStatus = "COMPLETED"
)

var AllStatuses = []ProjectStatus{
	NOT_INITIALIZED,
	IN_PROGRESS,
	COMPLETED,
}

func (s ProjectStatus) IsValid() bool {
	for _, v := range AllStatuses {
		if s == v {
			return true
		}
	}
	return false
}
