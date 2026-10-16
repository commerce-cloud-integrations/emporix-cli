package cli

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// importStatistics is the part of the import statistics shown in tables: the
// headline counters of the window.
type importStatistics struct {
	Summary struct {
		TotalRuns     json.Number `json:"totalRuns"`
		Succeeded     json.Number `json:"succeeded"`
		Partial       json.Number `json:"partial"`
		Failed        json.Number `json:"failed"`
		Cancelled     json.Number `json:"cancelled"`
		ActiveRuns    json.Number `json:"activeRuns"`
		RecordsRead   json.Number `json:"recordsRead"`
		FailedRecords json.Number `json:"failedRecords"`
	} `json:"summary"`
}

// jobGroup is the part of an import job group shown in tables.
type jobGroup struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	ConfigIDs []json.RawMessage `json:"configIds"` // counted only
	UpdatedAt string            `json:"updatedAt"`
}

// healthLevel is the health thresholds of one level, the tenant's or the
// built-in defaults (healthLevels).
type healthLevel struct {
	Level      string `json:"level"`
	Thresholds struct {
		FailureRatioAmber json.Number `json:"failureRatioAmber"`
		FailureRatioRed   json.Number `json:"failureRatioRed"`
		StaleAmberHours   json.Number `json:"staleAmberHours"`
		StaleRedHours     json.Number `json:"staleRedHours"`
		FailedRunIsRed    *bool       `json:"failedRunIsRed"`
	} `json:"thresholds"`
}

// importLicense is the tenant's import limits.
type importLicense struct {
	MaxRecordsPerEntity  json.Number `json:"maxRecordsPerEntity"`
	MaxConcurrentImports json.Number `json:"maxConcurrentImports"`
	BatchSize            json.Number `json:"batchSize"`
	Workers              json.Number `json:"workers"`
}

// importAnalyticsReads returns the reads of the import statistics, the job
// groups that scope them, the health thresholds and the import limits.
func importAnalyticsReads(a *app) []*cobra.Command {
	return []*cobra.Command{
		importReader(importStatisticsView()).command(a, &command{
			use:   "statistics",
			short: "Show the import statistics of the last 30 days",
			long: "Reads the import statistics of the last 30 days for the whole tenant: RUNS counts the runs\n" +
				"started, SUCCEEDED, PARTIAL, FAILED and CANCELLED how they ended, ACTIVE those queued or\n" +
				"running now, and RECORDS READ and RECORDS FAILED the records. -o json adds the rates, a daily\n" +
				"series, each stream's health, the failing streams, an error breakdown and the streams added\n" +
				"or removed. Another window or scope, e.g. one configuration (`epx imports job-groups` lists\n" +
				"sets of them): `epx api get /importtool/{tenant}/stats -p configId=ID\n" +
				"-p from=2026-09-01T00:00:00Z -p granularity=WEEK --scope " + importScope + "`.",
			req:   request{path: "/importtool/{tenant}/stats"},
			fetch: single{items: oneObject},
		}),
		importReader(view[jobGroup]{
			singular: "job group",
			header:   []string{columnID, columnName, "CONFIGURATIONS", columnModified},
			row: func(g jobGroup, _ language) []string {
				return []string{g.ID, g.Name, strconv.Itoa(len(g.ConfigIDs)), g.UpdatedAt}
			},
		}).command(a, &command{
			use:   "job-groups",
			short: "List the job groups that scope the import statistics",
			long: "Reads the tenant's job groups, named sets of import configurations that the import\n" +
				"statistics can be scoped to: CONFIGURATIONS counts them (-o json lists their IDs). The\n" +
				"service returns them whole, so they are read with one request.",
			req:   request{path: "/importtool/{tenant}/dashboard/job-groups"},
			fetch: listing{read: (*emporix.Client).GetUnpaged},
		}),
		importReader(healthLevelView()).command(a, &command{
			use:   "health-thresholds",
			short: "Show the thresholds that rate the health of import streams",
			long: "Reads the thresholds that rate an import stream GREEN, AMBER or RED in the statistics: the\n" +
				"tenant's, and the built-in defaults that each one it leaves empty falls back to. A stream is\n" +
				"AMBER or RED once this share of its records failed (FAILURE RATIO, 0 to 1) or its newest data\n" +
				"is this many hours old (STALE HOURS), and RED after a failed run when FAILED RUN RED is true.\n" +
				"A configuration or stream may set its own (in their -o json).",
			req:   request{path: "/importtool/{tenant}/settings/health-thresholds"},
			fetch: single{items: healthLevels},
		}),
		importReader(importLicenseView()).command(a, &command{
			use:   "license",
			short: "Show the tenant's import limits",
			long: "Reads the tenant's import limits: a run writes at most MAX RECORDS PER TYPE records per\n" +
				"target type (it skips the rest and ends PARTIAL), at most MAX CONCURRENT RUNS runs execute at\n" +
				"once (others wait, QUEUED), and BATCH SIZE records are written per request, by WORKERS\n" +
				"writers in parallel. Emporix support raises a limit on request.",
			req:   request{path: "/importtool/{tenant}/license"},
			fetch: single{items: oneObject},
		}),
	}
}

// importStatisticsView is the table of the headline of the import statistics.
func importStatisticsView() view[importStatistics] {
	return view[importStatistics]{
		singular: "import statistics",
		header: []string{
			"RUNS", "SUCCEEDED", "PARTIAL", columnFailed, "CANCELLED", columnActive, "RECORDS READ", "RECORDS FAILED",
		},
		row: func(s importStatistics, _ language) []string {
			sum := s.Summary
			return []string{reported(sum.TotalRuns), sum.Succeeded.String(), sum.Partial.String(), sum.Failed.String(),
				sum.Cancelled.String(), sum.ActiveRuns.String(), sum.RecordsRead.String(), sum.FailedRecords.String()}
		},
	}
}

// healthLevelView is the table of the health thresholds of the tenant and
// their built-in defaults.
func healthLevelView() view[healthLevel] {
	return view[healthLevel]{
		singular: "health thresholds",
		header: []string{
			"LEVEL", "FAILURE RATIO AMBER", "FAILURE RATIO RED", "STALE HOURS AMBER", "STALE HOURS RED",
			"FAILED RUN RED",
		},
		row: func(h healthLevel, _ language) []string {
			t := h.Thresholds
			return []string{h.Level, t.FailureRatioAmber.String(), t.FailureRatioRed.String(), t.StaleAmberHours.String(),
				t.StaleRedHours.String(), formatBool(t.FailedRunIsRed)}
		},
	}
}

// importLicenseView is the table of the tenant's import limits.
func importLicenseView() view[importLicense] {
	return view[importLicense]{
		singular: "import limits",
		header:   []string{"MAX RECORDS PER TYPE", "MAX CONCURRENT RUNS", "BATCH SIZE", "WORKERS"},
		row: func(l importLicense, _ language) []string {
			return []string{reported(l.MaxRecordsPerEntity), l.MaxConcurrentImports.String(), l.BatchSize.String(),
				l.Workers.String()}
		},
	}
}

// healthLevels returns the health thresholds of a response
// ({"thresholds": …, "builtInDefaults": …}) as two items (healthLevel): the
// tenant's, then the built-in defaults.
func healthLevels(body json.RawMessage) ([]json.RawMessage, error) {
	var settings *struct {
		Thresholds      json.RawMessage `json:"thresholds"`
		BuiltInDefaults json.RawMessage `json:"builtInDefaults"`
	}
	if err := json.Unmarshal(body, &settings); err != nil || settings == nil {
		return nil, errors.New("expected a JSON object")
	}
	type level struct {
		Level      string          `json:"level"`
		Thresholds json.RawMessage `json:"thresholds"` // null when missing
	}
	return marshalEach([]level{
		{Level: "tenant", Thresholds: settings.Thresholds},
		{Level: "built-in default", Thresholds: settings.BuiltInDefaults},
	})
}
