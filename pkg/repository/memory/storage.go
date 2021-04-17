package memory

import (
	"io"
	"log"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository"
)

// Clock provides Now method to enable mocking
type Clock interface {
	Now() time.Time
}

// Storage storage keeps data in memory
type Storage struct {
	Rand     io.Reader
	Clock    Clock
	comments []Comment
	//	worknotes []Comment
}

// AddComment saves the given asset to the repository and returns it's ID
func (m *Storage) AddComment(c comment.Comment) (string, error) {
	id, err := repository.GenerateUUID(m.Rand)
	if err != nil {
		log.Fatal(err)
	}

	createdBy := CreatedBy{}
	if c.CreatedBy != nil {
		createdBy.UUID = c.CreatedBy.UUID
		createdBy.Name = c.CreatedBy.Name
	}

	newC := Comment{
		ID:        id,
		Entity:    c.Entity,
		Text:      c.Text,
		CreatedBy: createdBy,
		CreatedAt: m.Clock.Now().Format(time.RFC3339),
	}
	m.comments = append(m.comments, newC)

	return id, nil
}

// GetComment returns a comment with the specified ID
func (m *Storage) GetComment(id string) (comment.Comment, error) {
	var c comment.Comment

	for i := range m.comments {

		if m.comments[i].ID == id {
			sc := m.comments[i]
			c.UUID = sc.ID
			c.Entity = sc.Entity
			c.Text = sc.Text
			c.ExternalID = sc.ExternalID
			c.CreatedAt = sc.CreatedAt
			if sc.CreatedBy.UUID != "" {
				createdBy := &comment.CreatedBy{
					UUID: sc.CreatedBy.UUID,
					Name: sc.CreatedBy.Name,
				}
				c.CreatedBy = createdBy
			}
			//createdBy :=
			//c.CreatedBy.UUID = m.comments[i].CreatedBy.UUID
			//c.CreatedBy.Name = m.comments[i].CreatedBy.Name

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

// QueryComments is not implemented
func (m *Storage) QueryComments(_ map[string]interface{}) (listing.QueryResult, error) {
	panic("not implemented")
}
