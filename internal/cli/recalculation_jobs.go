package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// recalculationJob is the part of a Product Service variant recalculation job
// shown in tables.
type recalculationJob struct {
	ID         string      `json:"id"`
	ParentRoot string      `json:"parentRoot"`
	Status     string      `json:"status"`
	Attempts   json.Number `json:"attempts"` // kept as sent
	CreatedAt  string      `json:"createdAt"`
	FinishedAt string      `json:"finishedAt"`
	Error      string      `json:"error"`
}

// newRecalculationJobsCommand reads the jobs that recalculate the variant
// trees of dynamic variant products. Emporix reads them only with the
// write-capable product.product_manage scope, so they are opt-in, and the
// reference documents no paging for them (it returns every job), so they are
// read with one request.
func newRecalculationJobsCommand(a *app) *cobra.Command {
	const columnError = "ERROR" // why a job failed: a long value
	return newResourceCommand(a, resource[recalculationJob]{
		use:   "recalculation-jobs",
		noun:  "variant recalculation jobs",
		under: productsUse,
		long: "Reads the jobs that rebuild the variant tree of a DYNAMIC_VARIANT product once a\n" +
			"recalculation was triggered, one per ROOT PRODUCT (an id). STATUS is PENDING, PROCESSING,\n" +
			"FINISHED, FAILED or FAILED_PERMANENT, ATTEMPTS counts processing attempts and ERROR says why a\n" +
			"job failed; -o json adds the product ids it was requested for and when processing started,\n" +
			"and leaves out each of processingStartedAt, finishedAt and error until it is set.\n" +
			"Emporix removes jobs about 30 days after they end. The jobs of one status: `epx api get\n" +
			"/product/{tenant}/products/recalculate/jobs -p status=FAILED --scope product.product_manage`.",
		path:   "/product/{tenant}/products/recalculate/jobs",
		scopes: []string{productManageScope},
		read:   (*emporix.Client).GetUnpaged,
		view: view[recalculationJob]{
			singular: "recalculation job",
			header:   []string{columnID, "ROOT PRODUCT", columnStatus, "ATTEMPTS", columnCreated, "FINISHED", columnError},
			long:     []string{columnError},
			row: func(j recalculationJob, _ language) []string {
				return []string{j.ID, j.ParentRoot, j.Status, j.Attempts.String(), j.CreatedAt, j.FinishedAt, j.Error}
			},
		},
	})
}
