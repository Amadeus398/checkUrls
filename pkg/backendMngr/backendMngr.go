package backendMngr

import (
	"CheckUrls/pkg/db"
	"CheckUrls/pkg/logging"
	"CheckUrls/pkg/repository/sites"
	statuses "CheckUrls/pkg/repository/status"
	"context"
	"database/sql"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"sync"
	"time"
)

const sqlLastCheckStatus = "SELECT  s.id AS site_id, s.url, s.frequency, st.date FROM sites s LEFT JOIN " +
	"(SELECT max(date) AS date, site_id FROM status GROUP BY site_id) st on s.id = st.site_id " +
	"WHERE s.deleted=$1 ORDER BY s.id DESC;"

type BackendManager struct {
	checks map[string]*check
	ctx    context.Context
	log    *logging.Loggers
	mux    sync.RWMutex
}

type check struct {
	site       *sites.Site
	tickCheck  *time.Ticker
	timerCheck *time.Timer
	stop       chan struct{}
}

func NewBackendManager(ctx context.Context) *BackendManager {
	logger := logging.NewLoggers("backendMngr", "newBackendManager")
	checkMap := make(map[string]*check)
	rows, cancel, err := db.ConnManager.Query(sqlLastCheckStatus, false)
	if err != nil {
		logger.ErrorLog().Err(err).Str("when", "sql request").Msg("failed sql query request")
		return nil
	}
	defer cancel()
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()
	for rows.Next() {
		lastCheck := check{
			site: &sites.Site{},
			stop: make(chan struct{}),
		}
		var lastDate sql.NullTime
		if err := rows.Scan(&lastCheck.site.Id, &lastCheck.site.Url, &lastCheck.site.Frequency, &lastDate); err != nil {
			log.Error().Err(err).Msg("unable to scan results")
			return nil
		}
		subDate := time.Now().Sub(lastDate.Time)
		switch {
		case !lastDate.Valid || time.Duration(lastCheck.site.Frequency) < subDate:
			go lastCheck.checkStatus()
			lastCheck.tickCheck = time.NewTicker(time.Duration(lastCheck.site.Frequency) * time.Second)
			go lastCheck.serve(ctx)
		case time.Duration(lastCheck.site.Frequency) > subDate:
			lastCheck.timerCheck = time.NewTimer((time.Duration(lastCheck.site.Frequency) - subDate) * time.Second)
			go lastCheck.serveOnce(ctx)
		}
		checkMap[lastCheck.site.Url] = &lastCheck
	}

	return &BackendManager{
		checks: checkMap,
		ctx:    ctx,
	}
}

func (m *BackendManager) CreateOrUpdate(site *sites.Site) {
	_, ok := m.checks[site.Url]
	if !ok {
		if err := m.registerSite(site); err != nil {
			m.log.ErrorLog().Err(err).Msg("unable to register site")
			return
		}
	}

	if err := m.updateSite(site); err != nil {
		m.log.ErrorLog().Err(err).Msg("unable to update site")
		return
	}
}

func (m *BackendManager) Delete(site *sites.Site) {
	m.checks[site.Url].close()
	delete(m.checks, site.Url)
}

func (m *BackendManager) registerSite(site *sites.Site) error {

	check := check{
		site:      site,
		tickCheck: time.NewTicker(time.Duration(site.Frequency) * time.Second),
		stop:      make(chan struct{}),
	}
	if site.Frequency == 0 {
		check.tickCheck = time.NewTicker(24 * time.Hour)
	}
	m.checks[site.Url] = &check
	go m.checks[site.Url].serve(m.ctx)

	return nil
}

func (m *BackendManager) updateSite(sites *sites.Site) error {
	m.checks[sites.Url].close()
	if err := m.registerSite(sites); err != nil {
		m.log.ErrorLog().Err(err).Msg("unable to update site")
		return err
	}
	return nil
}

func (c *check) close() {
	close(c.stop)
}

func (c *check) checkStatus() {
	logger := logging.NewLoggers("backendMngr", "checkStatus")
	logger.InfoLog().Msg("start check")
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 10 * time.Second,
	}
	stat := 0
	resp, err := client.Get(c.site.Url)
	if err != nil {
		logger.WarnLog().Err(err).Msg("unable to get response")
		stat = 600
	}
	date := timestamppb.Now()
	if stat != 600 {
		stat = resp.StatusCode
	}
	state := &statuses.State{
		Date:   date.AsTime(),
		Status: int64(stat),
		SiteId: c.site.Id,
	}
	if err := statuses.CreateStatus(state); err != nil {
		logger.ErrorLog().Err(err).Msg("unable to create status")
		return
	}
	logger.InfoLog().Str("when", "start check").Msg("done")
}

func (c *check) serveOnce(ctx context.Context) {
	defer c.timerCheck.Stop()
	select {
	case <-c.timerCheck.C:
		go c.checkStatus()
		c.tickCheck = time.NewTicker(time.Duration(c.site.Frequency) * time.Second)
		go c.serve(ctx)
		return
	case <-c.stop:
		return
	case <-ctx.Done():
		return
	}
}

func (c *check) serve(ctx context.Context) {
	defer c.tickCheck.Stop()
	for {
		select {
		case <-c.tickCheck.C:
			c.checkStatus()
		case <-c.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}
