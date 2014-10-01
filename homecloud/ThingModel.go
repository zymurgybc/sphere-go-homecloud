package homecloud

import (
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/redigo/redis"
)

type ThingModel struct {
	baseModel
}

func NewThingModel(conn redis.Conn) *ThingModel {
	return &ThingModel{baseModel{conn, "thing"}}
}

func (m *ThingModel) FetchByDeviceId(deviceId string) (*model.Thing, error) {
	device, err := deviceModel.Fetch(deviceId)
	if err != nil {
		return nil, err
	}

	if device.Thing == nil {
		return nil, nil
	}

	return m.Fetch(*device.Thing)
}

func (m *ThingModel) SetLocation(thingID string, roomID *string) error {

	var err error

	if roomID == nil {
		_, err = m.conn.Do("HDEL", "thing:"+thingID, "location")
	} else {
		_, err = m.conn.Do("HSET", "thing:"+thingID, "location", *roomID)
	}

	return err
}

func (m *ThingModel) Fetch(id string) (*model.Thing, error) {
	thing := &model.Thing{}

	if err := m.fetch(id, thing); err != nil {
		return nil, err
	}

	if thing.DeviceID != nil {
		device, err := deviceModel.Fetch(*thing.DeviceID)
		if err != nil {
			return nil, err
		}
		thing.Device = device
	}

	return thing, nil
}

func (m *ThingModel) FetchAll() (*[]*model.Thing, error) {

	ids, err := m.fetchAllIds()

	if err != nil {
		return nil, err
	}

	things := make([]*model.Thing, len(ids))

	for i, id := range ids {
		things[i], err = m.Fetch(id)
		if err != nil {
			return nil, err
		}
	}

	return &things, nil
}
