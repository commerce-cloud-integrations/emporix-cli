package cli

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// mediaManage is the scope of every write of media assets, as the reference
// names it (media.asset_manage_by_vendor is a vendor's alternative).
const mediaManage = "media.asset_manage"

// mediaWrites returns the writes of the media assets of assets: create and
// update, which send the JSON of a link or, with --upload, a form of a file
// and its JSON (assetForm), patch, delete and start-upload, which starts an
// upload straight to storage.
func mediaWrites(a *app, assets resource[mediaAsset]) []*cobra.Command {
	scopes := []string{mediaManage}
	unknown := "\nAn unknown ID is not found (exit 4)."
	return []*cobra.Command{
		(&write{
			use: createUse, short: "Create a media asset: a link, or a file to upload", method: http.MethodPost,
			long: "Creates a media asset from the JSON in --file and prints its id. A LINK asset points to a\n" +
				"file elsewhere, e.g. {\"type\": \"LINK\", \"access\": \"PUBLIC\", \"url\": \"https://example.com/a.pdf\"}.\n" +
				"A BLOB asset is a file that Emporix stores: --" + uploadFlag + " PATH sends the file in a\n" +
				"multipart/form-data form, with the JSON of --file as its body field, e.g.\n" +
				`{"type": "BLOB", "access": "PRIVATE", "details": {"filename": "a", "mimeType": "application/pdf"}}.` + "\n" +
				"Emporix takes a file of " + uploadLimitHelp + "; a large file may need a longer\n" +
				"--timeout, e.g. --timeout 5m, as it bounds the whole request.\n" + blobNeedsFileHelp + "\n" +
				"refIds assigns the asset to items, e.g. [{\"type\": \"PRODUCT\", \"id\": \"100200\"}]; an asset of a\n" +
				"PRODUCT or CATEGORY must be PUBLIC. An id that exists already is a conflict (exit 5).",
			example: "  epx media create --file link.json --dry-run\n" +
				"  epx media create --upload manual.pdf --file manual.json",
			req: request{path: assets.path}, scopes: scopes, body: assetForm(),
		}).command(a),
		(&write{
			use: assets.itemUse("update"), short: "Replace a media asset, or the file of one", method: http.MethodPut,
			long: "Replaces the media asset ID with the JSON in --file; its type and access cannot change.\n" +
				"The JSON of a LINK asset gives its url; a BLOB asset is replaced with its file: --" + uploadFlag +
				" PATH\nsends the new file in a multipart/form-data form, with the JSON as its body field\n" +
				"(" + uploadLimitHelp + "; a large file may need a longer\n" +
				"--timeout, e.g. --timeout 5m). `epx media patch` changes a BLOB asset without sending its\n" +
				"file again.\n" + blobNeedsFileHelp + unknown,
			req: request{path: assets.itemPath()}, scopes: scopes, body: assetForm(), ids: assets.itemIDs(),
		}).command(a),
		(&write{
			use: assets.itemUse("patch"), short: "Change members of a media asset", method: http.MethodPatch,
			long: "Changes the media asset ID with the JSON Patch (RFC 6902) in --file, an array of operations\n" +
				"such as [{\"op\": \"add\", \"path\": \"/refIds/-\", \"value\": {\"type\": \"PRODUCT\", \"id\": \"100200\"}}],\n" +
				"also of a BLOB asset, whose file it keeps; its type and access cannot change." + unknown,
			req: request{path: assets.itemPath()}, scopes: scopes, body: operationsFile(), ids: assets.itemIDs(),
		}).command(a),
		(&write{
			use: assets.itemUse("delete"), short: "Delete a media asset", method: http.MethodDelete,
			long: "Deletes the media asset ID, and the file of a BLOB asset. An unknown ID is no error: Emporix\n" +
				"answers as for a deleted asset.",
			req: request{path: assets.itemPath()}, scopes: scopes, body: noBody{}, ids: assets.itemIDs(),
		}).command(a),
		(&write{
			use: "start-upload", short: "Start an upload of a file straight to storage", method: http.MethodPost,
			long: "Starts an upload session: creates a PENDING BLOB asset from the JSON in --file, as create\n" +
				"takes it (\"uploadType\": \"form\" asks for a form instead of a PUT for a PRIVATE asset), and\n" +
				"prints where to send its file: the method, URL and headers or form fields of a request to\n" +
				"the storage (GCS for a PRIVATE asset, CLOUDINARY for a PUBLIC one), valid for 15 minutes\n" +
				"or an hour, so treat it as a secret. epx does not send the file there; once the storage\n" +
				"has it, the asset is READY. Emporix enables direct uploads per tenant; elsewhere it answers\n" +
				forbiddenRefusal + " and creates no asset.",
			req: request{path: assets.path + "/upload-session"}, scopes: scopes, body: &payload{},
			answers: emporix.Answers{ForbiddenIsRefusal: true},
		}).command(a),
	}
}

// blobNeedsFileHelp says in help that the JSON of a BLOB asset is not sent
// without a file to upload (blobNeedsFile).
const blobNeedsFileHelp = "Without --" + uploadFlag + ", the JSON of a BLOB asset is refused before anything " +
	"is sent, as Emporix\ntakes one only with its file."

// assetForm is the body of a write of a media asset: the JSON of --file,
// sent alone for a LINK asset or, with --upload, as the body field of a form
// that sends the file it names first, as the file field, for a BLOB asset;
// that of a BLOB asset is not sent alone (blobNeedsFile).
func assetForm() *formFile {
	return &formFile{
		refuseAlone: blobNeedsFile,
		form: form{uploadPart("file", "upload the file in `PATH` ("+uploadLimitHelp+") as the file of a "+
			"BLOB asset, in a multipart/form-data form with the JSON of --file; a large file may need a longer --timeout")},
		field: "body", usage: "JSON in `PATH`, sent byte for byte: the request body, or the body field of the form " +
			"of a file to upload",
	}
}

// blobNeedsFile returns why the JSON of a media asset cannot be sent without
// a file to upload: its type is "BLOB", and Emporix takes a BLOB asset only in
// a form with its file (it answers 400 Bad Request); nil for any other JSON.
// It reads type from a map, not a struct, whose keys Go matches regardless
// of case: only a member named exactly type counts, the last where it
// repeats, as a JSON parser's defaults read it.
func blobNeedsFile(asset json.RawMessage) error {
	var members map[string]json.RawMessage
	var kind string
	if json.Unmarshal(asset, &members) == nil && json.Unmarshal(members["type"], &kind) == nil && kind == "BLOB" {
		return errors.New("a BLOB asset is sent with its file: give --" + uploadFlag + " PATH (`epx media patch` " +
			"changes one without its file)")
	}
	return nil
}
