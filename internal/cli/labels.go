package cli

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
)

// label is the part of a Label Service label shown in tables.
type label struct {
	ID          string    `json:"id"`
	Name        localized `json:"name"` // a string in the reference; translations are accepted too
	Description localized `json:"description"`
	MediaID     string    `json:"mediaId"`
	Overlay     struct {
		IsTrue   *bool       `json:"isTrue"`
		Position json.Number `json:"position"` // kept as sent
	} `json:"overlay"`
}

// newLabelsCommand reads, creates, changes and deletes the tenant's labels, the
// badges shown on products. The Label Service has no tenant path segment: the
// tenant comes from the token. Its reference names the scope of label writes
// only among its OAuth scopes (label.label_manage, "needed to manage labels
// and label medias") and in the message of its 403 Forbidden answer. Emporix
// refuses a delete with that scope, though: it answers 403 naming
// label.label_delete, with which the delete succeeds.
func newLabelsCommand(a *app) *cobra.Command {
	labels := resource[label]{
		use:   "labels",
		short: "Read, create, change and delete labels",
		long: "Reads labels, the badges shown on products. OVERLAY is the label's overlay flag and\n" +
			"POSITION its overlay position, as sent; MEDIA is the id of its image (image URLs are only\n" +
			"in -o json), and DESCRIPTION is often HTML.",
		path:    "/label/labels",
		filters: listFilters{kinds: filterable},
		view: view[label]{
			singular:  "label",
			header:    []string{columnID, columnName, "OVERLAY", "POSITION", "MEDIA", columnDescription},
			localized: true,
			row: func(l label, lang language) []string {
				return []string{
					l.ID, l.Name.in(lang), formatBool(l.Overlay.IsTrue), l.Overlay.Position.String(), l.MediaID,
					l.Description.in(lang),
				}
			},
		},
	}
	manage := []string{"label.label_manage"}
	const answer = " It prints the label that Emporix answers with."
	return newResourceCommand(a, labels,
		(&write{
			use: createUse, short: "Create a label", method: http.MethodPost,
			long: "Creates a label from the JSON in --file, such as\n" +
				`{"name": "Free shipping", "description": "<p>Ships free</p>", "overlay": {"isTrue": true, ` +
				`"position": 1}},` + "\nand prints the label that Emporix answers with. Emporix generates its id " +
				"unless the JSON names one;\nan id that exists already is a conflict.",
			req: request{path: labels.path}, scopes: manage, body: &payload{},
		}).command(a),
		(&write{
			use: labels.itemUse("update"), short: "Replace a label", method: http.MethodPut,
			long: "Replaces the label ID with the JSON in --file, which must name its name and description.\n" +
				"An unknown ID is not found." + answer,
			req: request{path: labels.itemPath()}, scopes: manage, body: &payload{}, ids: labels.itemIDs(),
		}).command(a),
		(&write{
			use: labels.itemUse("patch"), short: "Change members of a label", method: http.MethodPatch,
			long: "Changes the members of the label ID that the JSON in --file names, which may be only\n" +
				"overlay; the others, such as name and description, keep their values. Unlike update,\n" +
				"Emporix does not check a metadata.version in it. An unknown ID is not found." + answer,
			req: request{path: labels.itemPath()}, scopes: manage, body: &payload{}, ids: labels.itemIDs(),
		}).command(a),
		itemDelete(labels, "Delete a label", "Deletes the label ID.", []string{"label.label_delete"}).command(a),
	)
}
