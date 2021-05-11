package backendMngr

import (
	"CheckUrls/pkg/db"
	"CheckUrls/pkg/logging"
	"CheckUrls/pkg/repository/sites"
	statuses "CheckUrls/pkg/repository/status"
	"context"
	"database/sql"
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
	logger.DebugLog().Msg("sql query get all sites with last check")
	rows, cancel, err := db.ConnManager.Query(sqlLastCheckStatus, false)
	if err != nil {
		logger.ErrorLog().Err(err).Str("when", "sql request").Msg("failed sql query request")
		return nil
	}
	defer cancel()
	defer func() {
		if err := rows.Close(); err != nil {
			logger.ErrorLog().Err(err).Str("when", "close rows").Msg("failed to close rows")
		}
	}()
	for rows.Next() {
		lastCheck := check{
			site: &sites.Site{},
			stop: make(chan struct{}),
		}
		var lastDate sql.NullTime
		if err := rows.Scan(&lastCheck.site.Id, &lastCheck.site.Url, &lastCheck.site.Frequency, &lastDate); err != nil {
			logger.ErrorLog().Err(err).Str("when", "scan rows").Msg("unable to scan results")
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
	m.log = logging.NewLoggers("backendNanager", "createOrUpdate")
	_, ok := m.checks[site.Url]
	m.log.DebugLog().Msg("create new site")
	if !ok {
		if err := m.registerSite(site); err != nil {
			m.log.ErrorLog().Err(err).Msg("unable to register site")
			return
		}
	}

	m.log.DebugLog().Msg("update site")
	if err := m.updateSite(site); err != nil {
		m.log.ErrorLog().Err(err).Msg("unable to update site")
		return
	}
}

func (m *BackendManager) Delete(site *sites.Site) {
	m.log = logging.NewLoggers("backendMngr", "delete")

	m.log.DebugLog().Msg("stop ticker")
	m.checks[site.Url].close()

	m.log.DebugLog().Msg("delete site from checkUrl")
	delete(m.checks, site.Url)
}

func (m *BackendManager) registerSite(site *sites.Site) error {
	m.log = logging.NewLoggers("backendMngr", "registerSite")

	m.log.DebugLog().Msg("filling out the site for verification")
	check := check{
		site:      site,
		tickCheck: time.NewTicker(time.Duration(site.Frequency) * time.Second),
		stop:      make(chan struct{}),
	}
	if site.Frequency == 0 {
		check.tickCheck = time.NewTicker(24 * time.Hour)
	}
	m.checks[site.Url] = &check

	m.log.DebugLog().Msg("starting the check")
	go m.checks[site.Url].serve(m.ctx)

	return nil
}

func (m *BackendManager) updateSite(sites *sites.Site) error {
	m.log = logging.NewLoggers("backendMngr", "updateSite")

	m.log.DebugLog().Msg("stop old ticker")
	m.checks[sites.Url].close()

	m.log.DebugLog().Msg("update site")
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

	logger.DebugLog().Msg("create client")
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 10 * time.Second,
	}
	stat := 0

	logger.DebugLog().Msg("send GET request")
	resp, err := client.Get(c.site.Url)
	if err != nil {
		logger.WarnLog().Err(err).Msg("unable to get response")
		stat = 600
	}

	logger.DebugLog().Msg("getting params of state")
	date := timestamppb.Now()
	if stat != 600 {
		stat = resp.StatusCode
	}
	state := &statuses.State{
		Date:   date.AsTime(),
		Status: int64(stat),
		SiteId: c.site.Id,
	}

	logger.DebugLog().Msg("create state")
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
