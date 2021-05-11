package sites

import (
	"CheckUrls/pkg/db"
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
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
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteFind, s.Url, true)
	if err != nil {
		log.Error().Err(err).Msg("unable to find site")
		return err
	}
	defer cancel()

	if err := row.Scan(&s.Id); err != nil {
		if err == sql.ErrNoRows {
			row, cancel, err = db.ConnManager.QueryRow(sqlSiteCreate, s.Url, s.Frequency, false)
			if err != nil {
				if err == db.ErrNothingDone {
					return ErrSitesNotFound
				}
				return err
			}
			defer cancel()
			if err := row.Scan(&s.Id); err != nil {
				return fmt.Errorf("ты вот тута попалась", err)
			}
			return nil
		}
		log.Error().Err(err).Msg("hui")
		return err
	}

	if err := db.ConnManager.Exec(sqlSiteUpdate, s.Url, s.Frequency, s.Id, false); err != nil {
		if err == db.ErrNothingDone {
			return ErrSitesNotFound
		}
		return err
	}

	return nil
}

func ReadSites(s *Site) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteRead, s.Id, false)
	if err != nil {
		if err == db.ErrNothingDone {
			return ErrSitesNotFound
		}
		return err
	}
	defer cancel()
	if err := row.Scan(&s.Id, &s.Url, &s.Frequency, &s.Deleted); err != nil {
		return err
	}
	return nil
}

func ReadAllSites() ([]*Site, error) {
	rows, cancel, err := db.ConnManager.Query(sqlSiteList, false)
	if err != nil {
		if err == db.ErrNothingDone {
			return nil, ErrSitesNotFound
		}
		return nil, err
	}
	defer cancel()
	defer func() {
		if err := rows.Close(); err != nil {
			err.Error()
		}
	}()
	list := make([]*Site, 0)
	for rows.Next() {
		s := new(Site)
		if err := rows.Scan(&s.Id, &s.Url, &s.Frequency, &s.Deleted); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func UpdateSites(s *Site) error {
	if err := db.ConnManager.Exec(sqlSiteUpdate, s.Url, s.Frequency, s.Id, false); err != nil {
		if err == db.ErrNothingDone {
			return ErrSitesNotFound
		}
		return err
	}
	return nil
}

func DeleteSites(s *Site) error {
	if err := db.ConnManager.Exec(sqlSiteDelete, s.Id, true); err != nil {
		if err == db.ErrNothingDone {
			return ErrSitesNotFound
		}
		return err
	}
	return nil
}
