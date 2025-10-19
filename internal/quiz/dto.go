package quiz

type QuizWithQuestionsDTO struct {
	Quiz      *Quiz           `json:"quiz"`
	Questions []*QuizQuestion `json:"questions"`
}
