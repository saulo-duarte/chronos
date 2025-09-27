package project

import "errors"

type UpdateProjectDTO struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status,omitempty"`
}

func (dto *UpdateProjectDTO) Validate() error {
	if dto.Title == "" {
		return errors.New("title cannot be empty")
	}
	if dto.Status != "" && !dto.Status.IsValid() {
		return errors.New("invalid project status")
	}
	return nil
}
