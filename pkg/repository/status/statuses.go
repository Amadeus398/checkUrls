package statuses

import (
	"CheckUrls/pkg/db"
	"CheckUrls/pkg/logging"
	"CheckUrls/pkg/proto"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	sqlCreateStatus = "INSERT INTO status (date, status_code, site_id) VALUES ($1, $2, $3);"
	sqlGetStatus    = "SELECT st.id, st.date, st.status_code, s.id AS site_id, s.url, s.frequency " +
		"FROM status st JOIN sites s ON s.id=st.site_id WHERE s.url=$1 ORDER BY st.date DESC LIMIT $2;"
)

var ErrStatusNotFound = fmt.Errorf("status not found")

type State struct {
	Id     int64
	Date   time.Time
	Status int64
	SiteId int64
}

func CreateStatus(status *State) error {
	log := logging.NewLoggers("statuses", "createStatus")
	log.DebugLog().Msg("processing the sql request")
	err := db.ConnManager.Exec(sqlCreateStatus, status.Date, status.Status, status.SiteId)
	if err != nil {
		if err == db.ErrNothingDone {
			err = ErrStatusNotFound
		}
		log.ErrorLog().Str("when", "processing the sql request").
			Err(err).Msg("unable to get row")
		return err
	}

	return nil
}

func ReadStatus(url string, count int64) (*proto.StatusResponse, error) {
	log := logging.NewLoggers("statuses", "readStatus")
	log.DebugLog().Msg("processing the sql request")
	rows, cancel, err := db.ConnManager.Query(sqlGetStatus, url, count)
	if err != nil {
		if err == db.ErrNothingDone {
			err = ErrStatusNotFound
		}
		log.ErrorLog().Err(err).Str("when", "processing the sql request").
			Msg("unable to get rows")
		return nil, err
	}
	defer cancel()
	defer func() {
		if err := rows.Close(); err != nil {
			log.ErrorLog().Err(err).Str("when", "close rows").Msg("unable to close rows")
			err.Error()
		}
	}()
	var frequency int64
	list := make([]*proto.State, 0)
	log.DebugLog().Msg("getting all rows")
	for rows.Next() {
		s := new(proto.State)
		var checkTime time.Time
		if err := rows.Scan(&s.Id, &checkTime, &s.Status, &s.SiteId, &url, &frequency); err != nil {
			log.ErrorLog().Err(err).Str("when", "getting all rows").Msg("unable to get rows")
			return nil, err
		}
		s.Date = timestamppb.New(checkTime)
		list = append(list, s)
	}

	return &proto.StatusResponse{
		Url:       url,
		Frequency: frequency,
		States:    list,
	}, nil

}
