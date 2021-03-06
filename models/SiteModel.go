package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/model"
)

type SiteModel struct {
	baseModel
}

func NewSiteModel() *SiteModel {
	return &SiteModel{
		baseModel: newBaseModel("site", model.Site{}),
	}
}

func (m *SiteModel) Fetch(id string) (*model.Site, error) {
	m.syncing.Wait()

	if id == "here" {
		id = config.MustString("siteId")
	}

	site := &model.Site{}

	if err := m.fetch(id, site, false); err != nil {
		return nil, err
	}

	return site, nil
}

func (m *SiteModel) FetchAll() (*[]*model.Site, error) {
	m.syncing.Wait()

	ids, err := m.fetchIds()

	if err != nil {
		return nil, err
	}

	sites := make([]*model.Site, len(ids))

	for i, id := range ids {
		sites[i], err = m.Fetch(id)
		if err != nil {
			return nil, err
		}
	}

	return &sites, nil
}

func (m *SiteModel) Create(site *model.Site) error {
	m.syncing.Wait()
	//defer m.sync()

	if site.ID == "here" {
		site.ID = config.MustString("siteId")
	}

	m.log.Debugf("Saving site %s", site.ID)

	updated, err := m.save(site.ID, site)

	m.log.Debugf("Site was updated? %t", updated)

	return err
}

func (m *SiteModel) Delete(id string) error {
	m.syncing.Wait()
	//defer m.sync()

	if id == "here" {
		id = config.MustString("siteId")
	}

	return m.delete(id)
}

func (m *SiteModel) Update(id string, site *model.Site) error {
	m.syncing.Wait()
	//defer m.sync()

	if id == "here" {
		id = config.MustString("siteId")
	}

	oldSite := &model.Site{}

	if err := m.fetch(id, oldSite, false); err != nil {
		return fmt.Errorf("Failed to fetch site (id:%s): %s", id, err)
	}

	oldSite.Name = site.Name
	oldSite.Type = site.Type
	oldSite.SitePreferences = site.SitePreferences

	if (oldSite.Latitude == nil || oldSite.Longitude == nil) || (*oldSite.Latitude != *site.Latitude || *oldSite.Longitude != *site.Longitude) {
		oldSite.Latitude = site.Latitude
		oldSite.Longitude = site.Longitude

		tz, err := getTimezone(*site.Latitude, *site.Longitude)
		if err != nil {
			return fmt.Errorf("Failed to get timezone: %s", err)
		}

		oldSite.TimeZoneID = tz.TimeZoneID
		oldSite.TimeZoneName = tz.TimeZoneName
		oldSite.TimeZoneOffset = tz.RawOffset // TODO: Not handling DST. Worth even having?
	}

	if _, err := m.save(id, oldSite); err != nil {
		return fmt.Errorf("Failed to update site (id:%s): %s", id, err)
	}

	return nil
}

type googleTimezone struct {
	DstOffset    *int    `json:"dstOffset,omitempty"`
	RawOffset    *int    `json:"rawOffset,omitempty"`
	Status       *string `json:"status,omitempty"`
	TimeZoneID   *string `json:"timeZoneId,omitempty"`
	TimeZoneName *string `json:"timeZoneName,omitempty"`
}

func getTimezone(latitude, longitude float64) (*googleTimezone, error) {

	// TODO: Send proper timestamp to get the dst... or...?
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/timezone/json?location=%f,%f&timestamp=1414645501", latitude, longitude)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Could not access schema " + resp.Status)
	}

	bodyBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tz googleTimezone
	err = json.Unmarshal(bodyBuff, &tz)
	if err != nil {
		return nil, err
	}

	if *tz.Status != "OK" {
		return nil, fmt.Errorf("Failed to get timezone: %s", *tz.Status)
	}

	/*

	   req := &geocode.Request{
	     Region:   "us",
	     Provider: geocode.GOOGLE,
	     Location: &geocode.Point{-33.86, 151.20},
	   }*/

	return &tz, nil
}
