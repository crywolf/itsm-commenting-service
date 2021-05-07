package listing_test

import (
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/memory"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommentService(t *testing.T) {
	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

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
	}

	clock := testutils.FixedClock{}
	mockStorage := &memory.Storage{
		Clock: clock,
	}

	lister := listing.NewService(mockStorage)
	assetType := "comment"

	id1, err := mockStorage.AddComment(c1, channelID, assetType)
	require.NoError(t, err)

	id2, err := mockStorage.AddComment(c2, channelID, assetType)
	require.NoError(t, err)

	com1, err := lister.GetComment(id1, channelID, assetType)
	require.NoError(t, err)
	assert.Equal(t, c1.Text, com1.Text)
	assert.Equal(t, c1.Entity, com1.Entity)
	assert.Equal(t, c1.CreatedBy, com1.CreatedBy)
	assert.Equal(t, clock.NowFormatted(), com1.CreatedAt)

	com2, err := lister.GetComment(id2, channelID, assetType)
	require.NoError(t, err)
	assert.Equal(t, c2.Text, com2.Text)
	assert.Equal(t, c2.Entity, com2.Entity)
	assert.Nil(t, c2.CreatedBy) // createdBy was not filled in
	assert.Equal(t, clock.NowFormatted(), com2.CreatedAt)

	com3, err := lister.GetComment("NonexistentID", channelID, assetType)
	require.EqualError(t, err, "record was not found")
	assert.Empty(t, com3)
}
