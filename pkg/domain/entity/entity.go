package entity

import "fmt"

// Entity represents some external entity reference in the form "<entity>:<UUID>"
// TODO implement as struct
type Entity string

// NewEntity return Entity type referencing some <entity> with <uuid>
func NewEntity(entity, uuid string) Entity {
	return Entity(fmt.Sprintf("%s:%s", entity, uuid))
}

//func (t Entity) MarshalJSON() ([]byte, error) {
//	return []byte(fmt.Sprintf("%s:%s", t.entity, t.uuid))
//}
