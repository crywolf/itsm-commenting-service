package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Entity represents some external entity reference in the form "<entity>:<UUID>"
type Entity struct {
	entity string
	uuid   string
}

// NewEntity returns Entity type referencing some <entity> with <uuid>
func NewEntity(entity, uuid string) Entity {
	return Entity{
		entity: entity,
		uuid:   uuid,
	}
}

// String return string representation of Entity type
func (e Entity) String() string {
	return fmt.Sprintf("%s:%s", strings.ToLower(e.entity), e.uuid)
}

// MarshalJSON returns Entity as the JSON encoding of Entity
func (e Entity) MarshalJSON() ([]byte, error) {
	if e.entity == "" || e.uuid == "" {
		return []byte(`""`), nil
	}

	return json.Marshal(e.String())
}

// UnmarshalJSON sets *Entity values from JSON data
func (e *Entity) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	fields := strings.Split(s, ":")
	if len(fields) != 2 {
		return nil
	}

	entityName := strings.ToLower(fields[0])
	entityUUID := fields[1]

	e.entity = entityName
	e.uuid = entityUUID

	return nil
}
