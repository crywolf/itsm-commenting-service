package memory

import (
	"context"
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
func (m *Storage) AddComment(ctx context.Context, c comment.Comment, channelID, assetType string, origin string) (*comment.Comment, error) {
	id, err := repository.GenerateUUID(m.Rand)
	if err != nil {
		log.Fatal(err)
	}

	createdBy := CreatedBy{}
	if c.CreatedBy != nil {
		createdBy.UUID = c.CreatedBy.UUID
		createdBy.Name = c.CreatedBy.Name
		createdBy.Surname = c.CreatedBy.Surname
	}

	newC := Comment{
		ID:        id,
		Entity:    c.Entity,
		Text:      c.Text,
		CreatedBy: createdBy,
		CreatedAt: m.Clock.Now().Format(time.RFC3339),
	}
	m.comments = append(m.comments, newC)

	//extend original comment with generated stuff
	c.UUID = id
	c.CreatedAt = newC.CreatedAt

	return &c, nil
}

// GetComment returns a comment with the specified ID
func (m *Storage) GetComment(ctx context.Context, id, channelID, assetType string) (comment.Comment, error) {
	var c comment.Comment

	for i := range m.comments {

		if m.comments[i].ID == id {
			sc := m.comments[i] // stored comment
			c.UUID = sc.ID
			c.Entity = sc.Entity
			c.Text = sc.Text
			c.ExternalID = sc.ExternalID

			if len(sc.ReadBy) > 0 {
				for _, rb := range sc.ReadBy {
					c.ReadBy = append(c.ReadBy, comment.ReadBy{
						Time: rb.Time,
						User: comment.UserInfo{
							UUID:           rb.User.UUID,
							Name:           rb.User.Name,
							Surname:        rb.User.Surname,
							OrgDisplayName: rb.User.OrgDisplayName,
							OrgName:        rb.User.OrgName,
						},
					})
				}
			}

			c.CreatedAt = sc.CreatedAt
			if sc.CreatedBy.UUID != "" {
				createdBy := &comment.UserInfo{
					UUID:    sc.CreatedBy.UUID,
					Name:    sc.CreatedBy.Name,
					Surname: sc.CreatedBy.Surname,
				}
				c.CreatedBy = createdBy
			}

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

// MarkAsReadByUser adds user info to read_by array to comment with specified ID
func (m *Storage) MarkAsReadByUser(ctx context.Context, id string, readBy comment.ReadBy, channelID, assetType string) (bool, error) {
	for i := range m.comments {
		if m.comments[i].ID == id {
			sc := m.comments[i] // stored comment
			if sc.ReadBy == nil {
				sc.ReadBy = make(ReadByList, 0)
			}

			for _, rb := range sc.ReadBy {
				if rb.User.UUID == readBy.User.UUID {
					return true, nil
				}
			}

			sc.ReadBy = append(sc.ReadBy, ReadBy{
				Time: readBy.Time,
				User: UserInfo{
					UUID:           readBy.User.UUID,
					Name:           readBy.User.Name,
					Surname:        readBy.User.Surname,
					OrgDisplayName: readBy.User.OrgDisplayName,
					OrgName:        readBy.User.OrgName,
				},
			})

			m.comments[i] = sc
		}
	}

	return false, nil
}

// QueryComments is not implemented
func (m *Storage) QueryComments(ctx context.Context, _ map[string]interface{}, _, _ string) (listing.QueryResult, error) {
	panic("not implemented")
}
