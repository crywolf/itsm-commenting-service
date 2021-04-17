package adding_test

import (
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/memory"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCommentService(t *testing.T) {
	c1 := comment.Comment{
		Text:   "Test 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
		CreatedBy: &comment.CreatedBy{
			UUID: "8540d943-8ccd-4ff1-8a08-0c3aa338c58e", Name: "Some user 1",
		},
	}

	c2 := comment.Comment{
		Text:   "Test 2",
		Entity: entity.NewEntity("incident", "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"),
		CreatedBy: &comment.CreatedBy{
			UUID: "12a0c65d-6efc-4346-b39b-a84cc0384c28", Name: "Some user 2",
		},
	}

	clock := testutils.FixedClock{}
	mockStorage := &memory.Storage{
		Clock: clock,
	}

	adder := adding.NewService(mockStorage)

	_, err := adder.AddComment(c1)
	require.NoError(t, err)

	_, err = adder.AddComment(c2)
	require.NoError(t, err)

	comments := mockStorage.GetAllComments()
	assert.Len(t, comments, 2)
}
