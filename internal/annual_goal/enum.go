package annual_goal

type AnnualGoalStatus string

const (
	AnnualGoalStatusActive    AnnualGoalStatus = "ACTIVE"
	AnnualGoalStatusCompleted AnnualGoalStatus = "COMPLETED"
	AnnualGoalStatusCanceled  AnnualGoalStatus = "CANCELED"
)
