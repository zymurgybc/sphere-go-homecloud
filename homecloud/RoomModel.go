package homecloud

import (
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/redigo/redis"
)

type RoomModel struct {
	baseModel
	log *logger.Logger
}

func NewRoomModel(conn redis.Conn) *RoomModel {
	return &RoomModel{baseModel{conn, "room"}, logger.GetLogger("RoomModel")}
}

func (m *RoomModel) Create(room *model.Room) error {
	return m.create(room.ID, room)
}

func (m *RoomModel) Fetch(id string) (*model.Room, error) {
	room := &model.Room{}

	if err := m.fetch(id, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (m *RoomModel) FetchAll() (*[]*model.Room, error) {

	ids, err := m.fetchAllIds()

	if err != nil {
		return nil, err
	}

	rooms := make([]*model.Room, len(ids))

	for i, id := range ids {
		rooms[i], err = m.Fetch(id)
		if err != nil {
			return nil, err
		}
	}

	return &rooms, nil
}

func (m *RoomModel) MoveThing(from *string, to *string, thing string) error {
	var err error

	if to != nil && from == nil {
		// we need to add it

		dest := "room:" + *to + ":things"

		m.log.Debugf("Adding thing %s to room %s", thing, *to)
		_, err = m.conn.Do("SADD", dest, thing)

	} else if from != nil && to == nil {
		// we need to remove it

		src := "room:" + *from + "s:things"

		m.log.Debugf("Removing thing %s from room %s", thing, *from)
		_, err = m.conn.Do("SREM", src, thing)

	} else {
		// need to move it

		src := "room:" + *from + "s:things"
		dest := "room:" + *to + ":things"

		m.log.Debugf("Moving thing %s from room %s to room %s", thing, *from, *to)

		result, err2 := m.conn.Do("SMOVE", src, dest, thing)
		if err2 != nil {
			return err2
		}

		if result == 0 {
			// Wasn"t in src, we should add it to dest
			_, err = m.conn.Do("SADD", dest, thing)
		}

	}

	return err

}
