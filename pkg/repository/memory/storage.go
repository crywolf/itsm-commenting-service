package memory

import (
	"log"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
)

// Storage storage keeps data in memory
type Storage struct {
	comments []Comment
	//	worknotes []Comment
}

// AddComment saves the given asset to the repository and returns it's ID
func (m *Storage) AddComment(c comment.Comment) (string, error) {
	id, err := repository.GenerateUUID()
	if err != nil {
		log.Fatal(err)
	}

	newC := Comment{
		ID:        id,
		Entity:    c.Entity,
		Text:      c.Text,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	m.comments = append(m.comments, newC)

	return id, nil
}

// GetComment returns a comment with the specified ID
func (m *Storage) GetComment(id string) (comment.Comment, error) {
	var c comment.Comment

	for i := range m.comments {

		if m.comments[i].ID == id {
			c.UUID = m.comments[i].ID
			c.Entity = m.comments[i].Entity
			c.Text = m.comments[i].Text
			c.ExternalID = m.comments[i].ExternalID
			c.CreatedAt = m.comments[i].CreatedAt

			return c, nil
		}
	}

	return c, ErrNotFound
}

// GetAllComments return all comments
func (m *Storage) GetAllComments() []comment.Comment {
	var comments []comment.Comment

	for i := range m.comments {

		c := comment.Comment{
			UUID:      m.comments[i].ID,
			Entity:    m.comments[i].Entity,
			Text:      m.comments[i].Text,
			CreatedAt: m.comments[i].CreatedAt,
		}

		comments = append(comments, c)
	}

	return comments
}
