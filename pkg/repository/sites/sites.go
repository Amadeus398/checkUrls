package sites

import (
	"CheckUrls/pkg/db"
	"CheckUrls/pkg/logging"
	"database/sql"
	"fmt"
)

const (
	sqlSiteCreate = "INSERT INTO sites (url, frequency, deleted) VALUES ($1, $2, $3) RETURNING id;"
	sqlSiteRead   = "SELECT * FROM sites WHERE id=$1 AND deleted=$2;"
	sqlSiteUpdate = "UPDATE sites SET url=$1, frequency=$2, deleted=$4 WHERE id=$3;"
	sqlSiteDelete = "UPDATE sites SET deleted=$2 WHERE id=$1;"
	sqlSiteList   = "SELECT * FROM sites WHERE deleted=$1;"
	sqlSiteFind   = "SELECT id FROM sites WHERE url=$1 AND deleted=$2;"
)

var ErrSitesNotFound = fmt.Errorf("sites not found")

type Site struct {
	Id        int64
	Url       string
	Frequency int64
	Deleted   bool
}

func CreateSites(s *Site) error {
	logger := logging.NewLoggers("sites", "createSite")
	logger.DebugLog().Msg("find the site")
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteFind, s.Url, true)
	if err != nil {
		logger.ErrorLog().Err(err).Str("when", "processing the sql request to find site").
			Msg("unable to find site")
		return err
	}
	defer cancel()

	if err := row.Scan(&s.Id); err != nil {
		if err == sql.ErrNoRows {
			logger.DebugLog().Str("when", "site not found").Msg("create new site")
			row, cancel, err = db.ConnManager.QueryRow(sqlSiteCreate, s.Url, s.Frequency, false)
			if err != nil {
				logger.ErrorLog().Err(err).Str("when", "processing sql request create site").
					Msg("unable to create site")
				return err
			}
			defer cancel()
			if err := row.Scan(&s.Id); err != nil {
				logger.ErrorLog().Err(err).Str("when", "scan site_id").
					Msg("failed to scan site_id")
				return err
			}
			return nil
		}
		logger.ErrorLog().Err(err).Msg("unable to create site")
		return err
	}

	logger.DebugLog().Str("when", "site found").Msg("processing sql request update site")
	if err := db.ConnManager.Exec(sqlSiteUpdate, s.Url, s.Frequency, s.Id, false); err != nil {
		if err == db.ErrNothingDone {
			logger.ErrorLog().Err(err).Str("when", "processing sql request update site").
				Str("when", "site not found").Msg("unable to update site")
			return ErrSitesNotFound
		}
		logger.ErrorLog().Err(err).Str("when", "processing sql request update site").
			Msg("unable to update site")
		return err
	}

	return nil
}

func ReadSites(s *Site) error {
	logger := logging.NewLoggers("sites", "readSite")

	logger.DebugLog().Msg("processing sql request read site")
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteRead, s.Id, false)
	if err != nil {
		logger.ErrorLog().Err(err).Str("when", "processing sql request read site").
			Msg("unable to read site")
		return err
	}
	defer cancel()
	logger.DebugLog().Msg("scan results")
	if err := row.Scan(&s.Id, &s.Url, &s.Frequency, &s.Deleted); err != nil {
		logger.ErrorLog().Err(err).Str("when", "scan results").Msg("unable to scan results")
		return err
	}
	return nil
}

func ReadAllSites() ([]*Site, error) {
	logger := logging.NewLoggers("sites", "readAllSites")

	logger.DebugLog().Msg("processing sql request read all sites")
	rows, cancel, err := db.ConnManager.Query(sqlSiteList, false)
	if err != nil {
		if err == db.ErrNothingDone {
			logger.ErrorLog().Err(err).Str("when", "processing sql request read all sites").
				Str("when", "site not found").Msg("unable to read all sites")
			return nil, ErrSitesNotFound
		}
		logger.ErrorLog().Err(err).Str("when", "processing sql request read all sites").
			Msg("unable to read all sites")
		return nil, err
	}
	defer cancel()
	defer func() {
		logger.DebugLog().Msg("close rows")
		if err := rows.Close(); err != nil {
			logger.ErrorLog().Err(err).Str("when", "close rows").Msg("unable to close rows")
			err.Error()
		}
	}()

	logger.DebugLog().Msg("getting list of sites")
	list := make([]*Site, 0)
	for rows.Next() {
		s := new(Site)
		logger.DebugLog().Str("when", "getting list of sites")
		if err := rows.Scan(&s.Id, &s.Url, &s.Frequency, &s.Deleted); err != nil {
			logger.ErrorLog().Err(err).Str("when", "scan results").
				Str("when", "getting list of sites").Msg("unable to scan results")
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func UpdateSites(s *Site) error {
	logger := logging.NewLoggers("sites", "updateSites")

	logger.DebugLog().Msg("processing sql request update site")
	if err := db.ConnManager.Exec(sqlSiteUpdate, s.Url, s.Frequency, s.Id, false); err != nil {
		if err == db.ErrNothingDone {
			logger.ErrorLog().Err(err).Str("when", "processing sql request update site").
				Str("when", "site not found").Msg("unable to update site")
			return ErrSitesNotFound
		}
		logger.ErrorLog().Err(err).Str("when", "processing sql request update site").
			Msg("unable to update site")
		return err
	}
	return nil
}

func DeleteSites(s *Site) error {
	logger := logging.NewLoggers("sites", "deleteSites")

	logger.DebugLog().Msg("processing sql request delete site")
	if err := db.ConnManager.Exec(sqlSiteDelete, s.Id, true); err != nil {
		if err == db.ErrNothingDone {
			logger.ErrorLog().Err(err).Str("when", "processing sql request delete site").
				Str("when", "site not found").Msg("unable to delete site")
			return ErrSitesNotFound
		}
		logger.ErrorLog().Err(err).Str("when", "processing sql request delete site").
			Msg("unable to delete site")
		return err
	}
	return nil
}
