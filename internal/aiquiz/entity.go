package aiquiz

type Question struct {
	Tema            string   `json:"tema"`
	Dificuldade     string   `json:"dificuldade"`
	Pergunta        string   `json:"pergunta"`
	Alternativas    []string `json:"alternativas"`
	RespostaCorreta string   `json:"resposta_correta"`
	Explicacao      string   `json:"explicacao"`
}

type QuestionRequest struct {
	Tema        string `json:"tema"`
	Dificuldade string `json:"dificuldade"`
	Quantidade  int    `json:"quantidade"`
}

type QuestionResponse struct {
	Questions []Question `json:"questions"`
}
