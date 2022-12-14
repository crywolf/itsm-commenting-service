package adding_test

import (
	"context"
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
		CreatedBy: &comment.UserInfo{
			UUID: "8540d943-8ccd-4ff1-8a08-0c3aa338c58e", Name: "Bob", Surname: "Martin",
		},
	}

	c2 := comment.Comment{
		Text:   "Test 2",
		Entity: entity.NewEntity("incident", "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"),
		CreatedBy: &comment.UserInfo{
			UUID: "12a0c65d-6efc-4346-b39b-a84cc0384c28", Name: "Alice", Surname: "Cooper",
		},
	}

	clock := testutils.FixedClock{}
	mockStorage := &memory.Storage{
		Clock: clock,
	}

	adder := adding.NewService(mockStorage)
	assetType := comment.AssetTypeComment

	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	ctx := context.Background()

	_, err := adder.AddComment(ctx, c1, channelID, assetType)
	require.NoError(t, err)

	_, err = adder.AddComment(ctx, c2, channelID, assetType)
	require.NoError(t, err)

	comments := mockStorage.GetAllComments()
	assert.Len(t, comments, 2)
}
