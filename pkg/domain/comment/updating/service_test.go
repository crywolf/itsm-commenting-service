package updating_test

import (
	"context"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/memory"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkAsReadByUserService(t *testing.T) {
	channelID := "e27ddcd0-0e1f-4bc5-93df-f6f04155beec"

	c1 := comment.Comment{
		Text:   "Test 1",
		Entity: entity.NewEntity("incident", "f49d5fd5-8da4-4779-b5ba-32e78aa2c444"),
		CreatedBy: &comment.UserInfo{
			UUID: "8540d943-8ccd-4ff1-8a08-0c3aa338c58e", Name: "Some user 1",
		},
	}

	c2 := comment.Comment{
		Text:   "Test 2",
		Entity: entity.NewEntity("incident", "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"),
		CreatedBy: &comment.UserInfo{
			UUID: "12a0c65d-6efc-4346-b39b-a84cc0384c28", Name: "Some user 2",
		},
	}

	ctx := context.Background()
	clock := testutils.FixedClock{}
	mockStorage := &memory.Storage{
		Clock: clock,
	}

	assetType := comment.AssetTypeComment

	adder := adding.NewService(mockStorage)

	com1ID, err := adder.AddComment(ctx, c1, channelID, assetType)
	require.NoError(t, err)

	com2ID, err := adder.AddComment(ctx, c2, channelID, assetType)
	require.NoError(t, err)

	updater := updating.NewService(mockStorage)

	readBy := comment.ReadBy{
		Time: "current timestamp",
		User: comment.UserInfo{
			UUID:           "439e2d19-8d50-405d-ad8e-cd33df344086",
			Name:           "Joe",
			Surname:        "Potato",
			OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
			OrgDisplayName: "Kompitech",
		},
	}
	alreadyRead, err := updater.MarkAsReadByUser(ctx, com1ID.UUID, readBy, channelID, assetType)
	require.NoError(t, err)
	assert.False(t, alreadyRead)

	alreadyRead, err = updater.MarkAsReadByUser(ctx, com1ID.UUID, readBy, channelID, assetType)
	require.NoError(t, err)
	assert.True(t, alreadyRead)

	readBy2 := comment.ReadBy{
		Time: "another timestamp",
		User: comment.UserInfo{
			UUID:           "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
			Name:           "Andy",
			Surname:        "Orange",
			OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
			OrgDisplayName: "Kompitech",
		},
	}
	alreadyRead, err = updater.MarkAsReadByUser(ctx, com1ID.UUID, readBy2, channelID, assetType)
	require.NoError(t, err)
	assert.False(t, alreadyRead)

	lister := listing.NewService(mockStorage)

	com1, err := lister.GetComment(ctx, com1ID.UUID, channelID, assetType)
	require.NoError(t, err)
	assert.NotNil(t, com1.ReadBy)
	assert.Len(t, com1.ReadBy, 2)
	assert.Equal(t, comment.ReadByList{readBy, readBy2}, com1.ReadBy)

	com2, err := lister.GetComment(ctx, com2ID.UUID, channelID, assetType)
	require.NoError(t, err)
	assert.Nil(t, com2.ReadBy)
}
