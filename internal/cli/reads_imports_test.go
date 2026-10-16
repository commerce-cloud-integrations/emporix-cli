package cli

import "slices"

// importReads returns the Import Service reads, every one opt-in: Emporix reads
// them only with importtool.import_trigger, which also starts runs.
func importReads() []readCase {
	const optIn = "--" + writeScopeFlag
	scopes := []string{importScope}
	run := `{"id":"r1","configId":"c1","status":"PARTIAL","mode":"DELTA","startedAt":"2026-07-27T02:00:00.000Z",
		"finishedAt":"2026-07-27T02:04:12.000Z","recordsRead":1250,"created":40,"updated":1180,"skipped":25,"failed":5,
		"deleted":0,"message":"Import completed with 5 record failures."}`
	return slices.Concat(importConfigurationReads(optIn, scopes, run), []readCase{
		{
			// A stream is read by its own ID; its settings are only in -o json.
			args:   []string{"imports", "streams", "get", "s 1", optIn},
			path:   "/importtool/acme1/streams/s%201",
			scopes: scopes,
			mode:   getWhole,
			body: `{"id":"s 1","configId":"c1","name":"Products","sourceEntity":"PRODUCT","targetWriter":"emporix.customEntity",
				"targetType":"product","mode":"STANDALONE","writeStrategy":"PATCH","enabled":true}`,
			want: "ID   NAME      SOURCE   TARGET TYPE  MODE        ENABLED\n" +
				"s 1  Products  PRODUCT  product      STANDALONE  true\n",
		},
		{
			args:   []string{"imports", "streams", "records", "s1", optIn},
			path:   "/importtool/acme1/data/streams/s1/records",
			scopes: scopes,
			mode:   getAllContentPages,
			body: `{"content":[{"id":"d1","streamId":"s1","targetType":"product","naturalKey":"100200","targetId":"t1",
				"fields":{"name":"Monkey wrench"},"outcome":"UPSERTED","importedAt":"2026-07-27T02:03:55.000Z"}],
				"totalElements":1,"totalPages":1,"number":0,"size":100}`,
			want: "ID  KEY     TYPE     OUTCOME   IMPORTED\n" +
				"d1  100200  product  UPSERTED  2026-07-27T02:03:55.000Z\n",
		},
		{
			// get shows the run; the progress of its streams is only in -o json there.
			args:   []string{"imports", "runs", "get", "r1", optIn},
			path:   "/importtool/acme1/runs/r1",
			scopes: scopes,
			mode:   getWhole,
			body:   `{"run":` + run + `,"streams":[{"streamName":"Products","status":"PARTIAL"}]}`,
			want: "ID  STATUS   MODE   STARTED                   FINISHED                  READ  FAILED\n" +
				"r1  PARTIAL  DELTA  2026-07-27T02:00:00.000Z  2026-07-27T02:04:12.000Z  1250  5\n",
		},
		{
			args:   []string{"imports", "runs", "streams", "r1", optIn},
			path:   "/importtool/acme1/runs/r1",
			scopes: scopes,
			mode:   getWhole,
			body: `{"run":` + run + `,"streams":[{"id":"p1","runId":"r1","streamId":"s1","streamName":"Products",
				"status":"PARTIAL","recordsRead":1250,"created":40,"updated":1180,"skipped":25,"failed":5,"deleted":0,
				"message":"5 records failed validation."},{"streamName":"Prices","status":"ABORTED",
				"message":"The mappings were never published."}]}`,
			want: "STREAM    STATUS   READ  CREATED  UPDATED  SKIPPED  DELETED  FAILED  MESSAGE\n" +
				"Products  PARTIAL  1250  40       1180     25       0        5       5 records failed validation.\n" +
				"Prices    ABORTED                                                    The mappings were never published.\n",
		},
		{
			args:   []string{"imports", "runs", "errors", "r1", optIn},
			path:   "/importtool/acme1/runs/r1/errors",
			scopes: scopes,
			mode:   getAllContentPages,
			body: `{"content":[{"id":"e1","runId":"r1","streamId":"s1","naturalKey":"100200",
				"targetEntity":"emporix.customEntity","stage":"MAPPING","errorCode":"REQUIRED_FIELD_MISSING",
				"message":"Required field 'name' is missing.","createdAt":"2026-07-27T02:03:41.000Z"}],
				"totalElements":1,"totalPages":1,"number":0,"size":100}`,
			want: "ID  KEY     STAGE    CODE                    MESSAGE\n" +
				"e1  100200  MAPPING  REQUIRED_FIELD_MISSING  Required field 'name' is missing.\n",
		},
		{
			// Every recorded row is asked for; the sample's size and cap are only in -o json.
			args:   []string{"imports", "runs", "diagnostics", "r1", optIn},
			path:   "/importtool/acme1/runs/r1/diagnostics",
			query:  "limit=50000",
			scopes: scopes,
			mode:   getWhole,
			body: `{"returned":2,"recorded":2,"cap":2000,"sampleTruncated":false,"note":"All recorded rows.",
				"rows":[{"streamName":"poLine","kind":"UNRESOLVED_PARENT","naturalKey":"ID:PO-4711","linkingField":"HEADER_ID",
				"linkingValue":"PO-4711","targetType":"PURCHASE_ORDER","waitingLines":7},
				{"streamName":"Products","kind":"REPEATED_KEY","naturalKey":"100200","occurrence":2}]}`,
			want: "KIND               STREAM    KEY         LINKING FIELD  LINKING VALUE  WAITING LINES  OCCURRENCE\n" +
				"UNRESOLVED_PARENT  poLine    ID:PO-4711  HEADER_ID      PO-4711        7              \n" +
				"REPEATED_KEY       Products  100200                                                   2\n",
		},
	}, importDataAndAnalyticsReads(optIn, scopes))
}

// importConfigurationReads returns the reads of import configurations: the
// configurations and their streams, schedule, stream order and runs.
func importConfigurationReads(optIn string, scopes []string, run string) []readCase {
	return []readCase{
		{
			// The configurations come whole; VERSION is kept as sent.
			args:   []string{"imports", "configurations", "list", optIn},
			path:   "/importtool/acme1/configs",
			scopes: scopes,
			mode:   getWhole,
			body: `[{"id":"c1","tenant":"acme1","name":"Product master","sourceConnId":"k1","enabled":true,"version":3,
				"createdBy":"jane.doe@example.com","updatedAt":"2026-07-20T08:00:00.000Z",
				"healthThresholds":{"staleRedHours":72}},{"id":"c2","name":"Prices","enabled":false}]`,
			want: "ID  NAME            ENABLED  VERSION  MODIFIED\n" +
				"c1  Product master  true     3        2026-07-20T08:00:00.000Z\n" +
				"c2  Prices          false             \n",
		},
		{
			args:   []string{"imports", "configurations", "get", "c 1", optIn},
			path:   "/importtool/acme1/configs/c%201",
			scopes: scopes,
			mode:   getWhole,
			body:   `{"id":"c 1","name":"Product master","enabled":true,"version":3.0}`,
			want:   "ID   NAME            ENABLED  VERSION  MODIFIED\nc 1  Product master  true     3.0      \n",
		},
		{
			args:   []string{"imports", "configurations", "streams", "c1", optIn},
			path:   "/importtool/acme1/configs/c1/streams",
			scopes: scopes,
			mode:   getWhole,
			body: `[{"id":"s1","configId":"c1","name":"Products","sourceEntity":"PRODUCT","targetType":"product",
				"mode":"STANDALONE","enabled":true},{"id":"s2","configId":"c1","name":"Lines","sourceEntity":"PO_LINE",
				"targetType":"purchaseOrder","mode":"COMPOSITE_CHILD","childStrategy":"EMBED","compositeParent":"Orders"}]`,
			want: "ID  NAME      SOURCE   TARGET TYPE    MODE             ENABLED\n" +
				"s1  Products  PRODUCT  product        STANDALONE       true\n" +
				"s2  Lines     PO_LINE  purchaseOrder  COMPOSITE_CHILD  \n",
		},
		{
			args:   []string{"imports", "configurations", "schedule", "c1", optIn},
			path:   "/importtool/acme1/configs/c1/schedule",
			scopes: scopes,
			mode:   getWhole,
			body: `{"id":"h1","configId":"c1","cron":"0 0 2 * * *","timezone":"Europe/Amsterdam","enabled":true,
				"nextFireAt":"2026-07-28T00:00:00.000Z"}`,
			want: "CRON         TIME ZONE         ENABLED  NEXT RUN\n" +
				"0 0 2 * * *  Europe/Amsterdam  true     2026-07-28T00:00:00.000Z\n",
		},
		{
			// The streams in the order they run, each after its prerequisites.
			args:   []string{"imports", "configurations", "stream-order", "c1", optIn},
			path:   "/importtool/acme1/configs/c1/stream-order",
			scopes: scopes,
			mode:   getWhole,
			body: `{"order":["vendor","purchaseOrder","purchaseOrderLine"],"prereqs":{"vendor":[],
				"purchaseOrder":["vendor"],"purchaseOrderLine":["purchaseOrder","vendor"]}}`,
			want: "STREAM             AFTER\n" +
				"vendor             \n" +
				"purchaseOrder      vendor\n" +
				"purchaseOrderLine  purchaseOrder, vendor\n",
		},
		{
			// Most recent first, in pages from 0; MODE marks a dry run, whose counters are 0.
			args:   []string{"imports", "configurations", "runs", "c1", optIn},
			path:   "/importtool/acme1/configs/c1/runs",
			scopes: scopes,
			mode:   getAllContentPages,
			body: `{"content":[{"id":"r2","configId":"c1","status":"QUEUED","mode":"FULL","dryRun":true,
				"startedAt":"2026-07-28T02:00:00.000Z","executionStartedAt":null,"finishedAt":null,"recordsRead":0,
				"failed":0,"queuePosition":1},` + run + `],"totalElements":2,"totalPages":1,"number":0,"size":100}`,
			want: "ID  STATUS   MODE          STARTED                   FINISHED                  READ  FAILED\n" +
				"r2  QUEUED   FULL dry run  2026-07-28T02:00:00.000Z                            0     0\n" +
				"r1  PARTIAL  DELTA         2026-07-27T02:00:00.000Z  2026-07-27T02:04:12.000Z  1250  5\n",
		},
	}
}

// importDataAndAnalyticsReads returns the reads of imported records and of the
// import statistics, job groups, health thresholds and limits.
func importDataAndAnalyticsReads(optIn string, scopes []string) []readCase {
	return []readCase{
		{
			args:   []string{"imports", "types", optIn},
			path:   "/importtool/acme1/data/types",
			scopes: scopes,
			mode:   getWhole,
			body:   `["product","category"]`,
			want:   "TYPE\nproduct\ncategory\n",
		},
		{
			// The type is a query parameter, not a path segment.
			args:   []string{"imports", "records", "product", optIn},
			path:   "/importtool/acme1/data/records",
			query:  "type=product",
			scopes: scopes,
			mode:   getAllContentPages,
			body: `{"content":[{"id":"d1","configId":"c1","streamId":"s1","runId":"r1","targetType":"product",
				"naturalKey":"100200","targetId":"t1","outcome":"UPSERTED","importedAt":"2026-07-27T02:03:55.000Z"},
				{"id":"d2","targetType":"product","naturalKey":"100201","outcome":"FAILED"}],
				"totalElements":2,"totalPages":1,"number":0,"size":100}`,
			want: "ID  KEY     TYPE     OUTCOME   IMPORTED\n" +
				"d1  100200  product  UPSERTED  2026-07-27T02:03:55.000Z\n" +
				"d2  100201  product  FAILED    \n",
		},
		{
			// The headline counters of the last 30 days; the rest is only in -o json.
			args:   []string{"imports", "statistics", optIn},
			path:   "/importtool/acme1/stats",
			scopes: scopes,
			mode:   getWhole,
			body: `{"summary":{"totalRuns":42,"succeeded":38,"partial":3,"failed":1,"cancelled":0,"recordsRead":125000,
				"failedRecords":130,"successRate":0.905,"activeRuns":1},"series":[],"streamHealth":[],
				"topFailingStreams":[],"errorBreakdown":[],"streamChanges":{"added":2,"removed":1}}`,
			want: "RUNS  SUCCEEDED  PARTIAL  FAILED  CANCELLED  ACTIVE  RECORDS READ  RECORDS FAILED\n" +
				"42    38         3        1       0          1       125000        130\n",
		},
		{
			args:   []string{"imports", "job-groups", optIn},
			path:   "/importtool/acme1/dashboard/job-groups",
			scopes: scopes,
			mode:   getWhole,
			body: `[{"id":"g1","name":"Nightly master data","configIds":["c1","c2"],"createdAt":"2026-07-01T10:15:30.000Z",
				"updatedAt":"2026-07-20T08:00:00.000Z"},{"id":"g2","name":"Empty"}]`,
			want: "ID  NAME                 CONFIGURATIONS  MODIFIED\n" +
				"g1  Nightly master data  2               2026-07-20T08:00:00.000Z\n" +
				"g2  Empty                0               \n",
		},
		{
			// The tenant's thresholds, empty where it falls back to the built-in defaults.
			args:   []string{"imports", "health-thresholds", optIn},
			path:   "/importtool/acme1/settings/health-thresholds",
			scopes: scopes,
			mode:   getWhole,
			body: `{"thresholds":{"failureRatioAmber":0.05,"staleRedHours":72},"builtInDefaults":{"failureRatioAmber":0.01,
				"failureRatioRed":0.1,"staleAmberHours":48,"staleRedHours":168,"failedRunIsRed":true}}`,
			want: "LEVEL             FAILURE RATIO AMBER  FAILURE RATIO RED  STALE HOURS AMBER  STALE HOURS RED  " +
				"FAILED RUN RED\n" +
				"tenant            0.05                                                       72               \n" +
				"built-in default  0.01                 0.1                48                 168              true\n",
		},
		{
			args:   []string{"imports", "license", optIn},
			path:   "/importtool/acme1/license",
			scopes: scopes,
			mode:   getWhole,
			body:   `{"maxRecordsPerEntity":10000,"maxConcurrentImports":2,"batchSize":100,"workers":1}`,
			want: "MAX RECORDS PER TYPE  MAX CONCURRENT RUNS  BATCH SIZE  WORKERS\n" +
				"10000                 2                    100         1\n",
		},
	}
}
