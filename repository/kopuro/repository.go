package kopuro

import (
	"dang/model"
)

type Repository struct {
	kc kopuroClient
}

func NewRepository(kopuroBaseURL string) Repository {
	kc := newKopuroClient(kopuroBaseURL)
	return Repository{kc}
}

func (repo Repository) InsertDecision(decision model.Decision) error {
	return repo.kc.writeJSONFile(decision.Name, decision)
}