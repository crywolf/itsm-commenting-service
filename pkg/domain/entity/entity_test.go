package entity_test

import (
	"encoding/json"
	"testing"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/stretchr/testify/require"
)

func TestEntity_MarshalJSON(t *testing.T) {
	e := entity.NewEntity("incident", "a0642910-df26-4415-8f38-5c8663d90497")

	data, err := e.MarshalJSON()
	require.Empty(t, err)
	require.Equal(t, `"incident:a0642910-df26-4415-8f38-5c8663d90497"`, string(data))

	// Entity name upper case
	eUpper := entity.NewEntity("REQUEST", "2f572910-df26-4415-8f38-5c8663d90497")

	dataLower, err := eUpper.MarshalJSON()
	require.Empty(t, err)
	require.Equal(t, `"request:2f572910-df26-4415-8f38-5c8663d90497"`, string(dataLower))

}

func TestEntityMarshalling(t *testing.T) {
	e := entity.NewEntity("incident", "79ee4c40-e86a-4df4-899d-a26ac5924058")
	c := comment.Comment{
		Entity: e,
	}

	JSONData, err := json.Marshal(c)
	require.Empty(t, err)
	require.Equal(t,
		`{"entity":"incident:79ee4c40-e86a-4df4-899d-a26ac5924058"}`,
		string(JSONData))
}

func TestEmptyEntityMarshalling(t *testing.T) {
	e := entity.Entity{}

	data, err := json.Marshal(e)
	require.Empty(t, err)
	require.Equal(t, `""`, string(data))
}

func TestEntity_String(t *testing.T) {
	e := entity.NewEntity("incident", "a0642910-df26-4415-8f38-5c8663d90497")

	data := e.String()
	require.Equal(t, `incident:a0642910-df26-4415-8f38-5c8663d90497`, data)
}

func TestEntity_UnmarshalJSON(t *testing.T) {
	JSONData := []byte(`{"entity":"incident:79ee4c40-e86a-4df4-899d-a26ac5924058","text":""}`)
	c := comment.Comment{}

	err := json.Unmarshal(JSONData, &c)
	require.Empty(t, err)

	e := entity.NewEntity("incident", "79ee4c40-e86a-4df4-899d-a26ac5924058")
	require.Equal(t, e, c.Entity)
}
