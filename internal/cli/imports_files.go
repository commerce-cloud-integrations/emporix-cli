package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

// The file queue of a file-based import stream: the files that it waits to
// import, in the order it reads them, and those that it dealt with, which
// files list reads; the most bytes an upload takes, which files limits
// reads; and the writes that upload a file, set the load order and remove a
// waiting file. Emporix allows each only with importScope.

// filesUse is the name of the group of a stream's file queue.
const filesUse = "files"

// filesSegment is the path of a stream's file queue below the stream.
const filesSegment = "files"

// maxFileHistory is the most files of a stream's history that the Import
// Service lists (historyLimit).
const maxFileHistory = 500

// importFile is the part of a file of a stream's file queue shown in tables.
type importFile struct {
	ID          string      `json:"id"`
	FileName    string      `json:"fileName"`
	Size        json.Number `json:"size"`
	Status      string      `json:"status"`
	QueuedAt    string      `json:"queuedAt"`
	ProcessedAt string      `json:"processedAt"`
	RecordsRead json.Number `json:"recordsRead"`
	Message     string      `json:"message"`
}

// queuedFile is a file of a stream's file queue as fileQueueItems lists it:
// the file, F, and its place in the load order among all the files waiting,
// e.g. "1 of 3", for a waiting one; none for one of the history. The shape
// keeps the file as sent (json.RawMessage), a table row decodes it
// (importFile).
type queuedFile[F any] struct {
	Order string `json:"order"`
	File  F      `json:"file"`
}

// fileLimits is what the environment takes in an upload to a stream: the
// most bytes, whether the API gateway in front of the service sets that
// limit, and whether it can store an uploaded file at all.
type fileLimits struct {
	MaxUploadBytes  json.Number `json:"maxUploadBytes"`
	GatewayLimited  *bool       `json:"gatewayLimited"`
	UploadSupported *bool       `json:"uploadSupported"`
}

// fileModeShownBy returns where a stream of streams shows its fileMode, as
// help quotes it: the SOURCE of its get, e.g. "SOURCE of `epx imports streams
// get ID`".
func fileModeShownBy(streams resource[importStream]) string {
	return "SOURCE of `epx " + streams.getLine() + "`"
}

// fileModesHelp says how a stream's run reads its waiting files, by the
// stream's fileMode, in the help of the file queue and of its list, where
// shownBy is where a stream shows its fileMode (fileModeShownBy).
func fileModesHelp(shownBy string) string {
	return "A file-based stream reads its files by its fileMode, SNAPSHOT by default. A SNAPSHOT stream\n" +
		"takes each file as the whole picture: its run reads only the most recently arrived waiting\n" +
		"file and may report the older ones superseded, in their MESSAGE. An INCREMENTAL stream takes\n" +
		"each file as a change set: its run reads them all, in load order, a later file's values\n" +
		"winning on a key that an earlier one carries too. A stream without fileMode reads no files\n" +
		"(" + shownBy + " shows it)."
}

// newImportFilesCommand returns the group of the file queue of a stream of
// streams, nested in a group of its own (filesOf): list and limits, which
// read it, and upload, order and delete.
func newImportFilesCommand(a *app, streams resource[importStream]) *cobra.Command {
	filesOf, shownBy := streams.nested(filesUse), fileModeShownBy(streams)
	modes := fileModesHelp(shownBy)
	files := streamFiles(filesOf, modes)
	return newGroup(filesOf.use, "Read and change the files that a file-based stream imports",
		"Reads the file queue of a file-based import stream: the files that it waits to import, in\n"+
			"the order that its next run reads them, and those it dealt with; uploads a file to it, sets\n"+
			"the order of the waiting files and removes one. Nothing is imported until a run starts\n"+
			"(`epx imports configurations trigger`).\n"+modes,
		childCommand(a, filesOf, files),
		childCommand(a, filesOf, fileUploadLimits(filesOf)),
		fileUpload(filesOf, files, shownBy).command(a),
		fileOrder(filesOf, files).command(a),
		fileDelete(filesOf, files).command(a))
}

// queueLine returns the command line of the command verb of the file queue
// of a stream of filesOf (nested), without its arguments, as help quotes it,
// e.g. "`epx imports streams files order`".
func queueLine(filesOf resource[importStream], verb string) string {
	return "`epx " + filesOf.commandLine(verb, request{}) + "`"
}

// queueList returns the command line of the list of files, the file queue of
// a stream of filesOf (nested), with its argument, as help quotes it, e.g.
// "`epx imports streams files list STREAM_ID`".
func queueList(filesOf resource[importStream], files child[queuedFile[importFile]]) string {
	return "`epx " + childListLine(filesOf, files) + "`"
}

// streamFiles declares the file queue of a stream of filesOf (nested), which
// list reads; modes says how a stream reads its files (fileModesHelp).
func streamFiles(filesOf resource[importStream], modes string) child[queuedFile[importFile]] {
	return child[queuedFile[importFile]]{
		use:   listUse,
		sub:   filesSegment,
		short: "List the files that a stream waits to import and those it dealt with",
		long: "Reads the file queue of a file-based stream. First come the files waiting, in the order\n" +
			"that its next run reads them: LOAD ORDER is a file's place among all the files waiting,\n" +
			"e.g. 1 of 3.\n" + modes + "\n" +
			"Emporix lists up to 2000 waiting files: where more wait, the last LOAD ORDER listed ends\n" +
			"below the total, e.g. 2000 of 2500, and " + queueLine(filesOf, "order") + " cannot reorder\n" +
			"them. Then comes the history, newest first and without LOAD ORDER: up to --history-limit\n" +
			"files that a run claimed (STATUS IN_PROGRESS), read to the end (PROCESSED, also with\n" +
			"records that failed, as MESSAGE says) or could not import (FAILED: it can be uploaded\n" +
			"again). SIZE is in bytes, QUEUED when the file was uploaded, PROCESSED when its import\n" +
			"ended and READ the records read from it. A file's load rank and run are only in -o json.",
		fetch: &paramRead{
			params: queryFlags{optional: []paramFlag{{
				param: "historyLimit", flag: "history-limit", check: fileHistoryLimit,
				usage: "the most `FILES` of the history to list, newest first: 1 to " + strconv.Itoa(maxFileHistory) +
					" (Emporix: 100 without it)",
			}}},
			fetch: single{items: fileQueueItems},
		},
		view: view[queuedFile[importFile]]{
			singular: "stream file",
			header: []string{
				columnID, "LOAD ORDER", columnName, columnStatus, "SIZE", "QUEUED", "PROCESSED", columnRead, columnMessage,
			},
			long: []string{columnMessage},
			row: func(f queuedFile[importFile], _ language) []string {
				file := f.File
				return []string{file.ID, f.Order, file.FileName, file.Status, file.Size.String(), file.QueuedAt,
					file.ProcessedAt, file.RecordsRead.String(), file.Message}
			},
		},
	}
}

// fileUploadLimits declares what the environment takes in an upload to a
// stream of filesOf (nested), which limits reads.
func fileUploadLimits(filesOf resource[importStream]) child[fileLimits] {
	return child[fileLimits]{
		use:   "limits",
		sub:   filesSegment + "/limits",
		short: "Show the most bytes that an upload to a stream takes",
		long: "Reads the largest file that the environment takes in " + queueLine(filesOf, "upload") + ":\n" +
			"MAX UPLOAD BYTES, which the API gateway in front of the Import Service sets where GATEWAY\n" +
			"LIMITED is true (it refuses a larger file with its own 413 before the service sees it), and\n" +
			"UPLOAD SUPPORTED, false where the environment has nowhere to store an uploaded file (an\n" +
			"upload is then answered 503 Service Unavailable). A larger file can be published at an\n" +
			"HTTP(S) URL that the stream's connection reads instead, a change that needs\n" +
			"importtool.import_manage, which epx does not request.",
		fetch: single{items: oneObject},
		view: view[fileLimits]{
			singular: "file upload limits",
			header:   []string{"MAX UPLOAD BYTES", "GATEWAY LIMITED", "UPLOAD SUPPORTED"},
			row: func(l fileLimits, _ language) []string {
				return []string{reported(l.MaxUploadBytes), formatBool(l.GatewayLimited), formatBool(l.UploadSupported)}
			},
		},
	}
}

// uploadConflict says what a 409 Conflict of an upload means and how to
// recover, where list is the stream's file list, as help quotes it
// (queueList), and shownBy where the stream shows its fileMode
// (fileModeShownBy).
func uploadConflict(list, shownBy string) string {
	return "the stream imported or queued this content already, reads no files, or has a run in flight that " +
		"reads what it would replace: " + list + " shows the files queued and the latest it dealt with " +
		"(--history-limit, up to " + strconv.Itoa(maxFileHistory) + "), by name, while Emporix matches content " +
		"whatever the name; a stream without fileMode (" + shownBy + " shows it) reads no files and takes none; " +
		"otherwise wait until the run ends"
}

// fileOrderConflict says what a 409 Conflict of an order of a stream's
// waiting files means and how to recover, where list is the stream's file
// list as help quotes it (queueList).
func fileOrderConflict(list string) string {
	return "the stream's job has a run in flight, or more files wait than one request orders: " + list +
		" shows a last LOAD ORDER below the total where more wait, and those cannot be reordered until fewer " +
		"wait; otherwise wait until the run ends"
}

// fileDeleteConflict says what a 409 Conflict of a delete of a waiting file
// means and how to recover, where list is the stream's file list as help
// quotes it (queueList).
func fileDeleteConflict(list string) string {
	return "the file no longer waits, as a run claimed or imported it, or the stream's job has a run in flight: " +
		list + " shows its STATUS; otherwise wait until the run ends"
}

// fileHistoryLimit refuses a history limit that is not a whole number from 1
// to maxFileHistory, which Emporix would change to the nearest of those
// (paramFlag.check).
func fileHistoryLimit(value string) error {
	if n, err := strconv.ParseUint(value, 10, 63); err != nil || n == 0 || n > maxFileHistory {
		return fmt.Errorf("use a whole number from 1 to %d", maxFileHistory)
	}
	return nil
}

// fileQueueItems returns the files of a file queue response ({"pending":
// [files], "history": [files], "pendingTotal": n, "pendingTruncated": b}) as
// items (queuedFile): the waiting files in their load order, each with its
// place among all the files waiting, then the history as sent.
func fileQueueItems(body json.RawMessage) ([]json.RawMessage, error) {
	var queue *struct {
		Pending          []json.RawMessage `json:"pending"`
		History          []json.RawMessage `json:"history"`
		PendingTotal     *int64            `json:"pendingTotal"`
		PendingTruncated bool              `json:"pendingTruncated"`
	}
	if err := json.Unmarshal(body, &queue); err != nil || queue == nil || queue.Pending == nil ||
		queue.History == nil {
		return nil, errors.New("expected an object with the pending files and the history, and an integer " +
			"pendingTotal and a boolean pendingTruncated, if any")
	}
	total := waitingTotal(len(queue.Pending), queue.PendingTotal, queue.PendingTruncated)
	files := make([]queuedFile[json.RawMessage], 0, len(queue.Pending)+len(queue.History))
	for i, file := range queue.Pending {
		files = append(files, queuedFile[json.RawMessage]{Order: strconv.Itoa(i+1) + " of " + total, File: file})
	}
	for _, file := range queue.History {
		files = append(files, queuedFile[json.RawMessage]{File: file})
	}
	return marshalEach(files)
}

// waitingTotal returns how many files wait, of which a file queue lists
// listed: its pendingTotal, where it sends one; else, where it says that it
// lists only some (pendingTruncated), more than those; else those.
func waitingTotal(listed int, total *int64, truncated bool) string {
	switch {
	case total != nil:
		return strconv.FormatInt(*total, 10)
	case truncated:
		return "more than " + strconv.Itoa(listed)
	default:
		return strconv.Itoa(listed)
	}
}

// fileUpload uploads a file to files, the file queue of a stream of filesOf
// (nested); shownBy is where a stream shows its fileMode (fileModeShownBy).
func fileUpload(filesOf resource[importStream], files child[queuedFile[importFile]], shownBy string) *write {
	stream, limits, list := filesOf.argument(), queueLine(filesOf, "limits"), queueList(filesOf, files)
	return &write{
		use: "upload " + stream, short: "Upload a file for a stream to import", method: http.MethodPost,
		long: "Uploads the delimited file (CSV or a variant) that --" + uploadFlag + " names to the queue of the\n" +
			"file-based stream " + stream + ", which imports it on its next run: nothing is imported now.\n" +
			"The file's name is sent with it and tells Emporix a compression, such as .gz. Emporix\n" +
			"answers with the file queued, whose id " + queueLine(filesOf, "order") + " and\n" +
			queueLine(filesOf, deleteUse) + " take. Which waiting files a run reads depends on the\n" +
			"stream's fileMode, as " + list + " says.\n" +
			"Emporix knows a file by its content, whatever its name, and never imports it twice: content\n" +
			"that the stream imported or queued already is a conflict (exit 5), as is an upload to a\n" +
			"stream that reads no files or one that would replace content that a run in flight reads; a\n" +
			"file that FAILED is taken again.\n" +
			"epx sends a file of " + uploadLimitHelp + ", and the environment may take less\n" +
			"(" + limits + "): Emporix answers a larger one with\n" +
			"413 Request Entity Too Large (exit 1). A large file may need a longer --timeout, e.g.\n" +
			"--timeout 5m, as it bounds the whole request.\n" +
			"A 503 Service Unavailable (exit 1) comes from an environment that cannot store uploads\n" +
			"(UPLOAD SUPPORTED false in " + limits + "), which queued nothing, or from\n" +
			"a gateway that failed, after which the file may have been queued: check\n" +
			list + " before you send it again; if it was\n" +
			"queued, a resend is a conflict (exit 5), not a second import.\n" +
			filesOf.argumentSentence(notFound),
		example: "  epx imports streams files upload 3fa85f64-5717-4562-b3fc-2c963f66afa6 --upload vendors.csv",
		req:     request{path: files.below(filesOf.path)}, scopes: importScopes(),
		body: &formBody{form: form{uploadPart("file", "the delimited file in `PATH` to queue ("+uploadLimitHelp+
			"), sent with its name").require()}},
		conflict: uploadConflict(list, shownBy),
	}
}

// fileOrder sets the order in which a stream of filesOf (nested) reads the
// waiting files of files, its file queue.
func fileOrder(filesOf resource[importStream], files child[queuedFile[importFile]]) *write {
	stream, list := filesOf.argument(), queueList(filesOf, files)
	return &write{
		use: "order " + stream, short: "Set the order in which a stream imports its waiting files",
		method: http.MethodPut,
		long: "Sets the order in which the stream " + stream + " reads its waiting files: --id names each one,\n" +
			"the first to load first, as " + list + " shows its ID;\n" +
			"--file gives the JSON array of their ids instead. The order matters only for an INCREMENTAL\n" +
			"stream, as that list says. Emporix takes every waiting file, no more and no fewer\n" +
			"(otherwise 400 Bad Request, exit 1), and answers with the waiting files in their new order.\n" +
			"A stream whose job has a run in flight, or with more waiting files than the list shows, is\n" +
			"a conflict (exit 5).\n" +
			filesOf.argumentSentence(notFound),
		example: "  epx imports streams files order 3fa85f64-5717-4562-b3fc-2c963f66afa6 " +
			"--id 9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d --id c56a4180-65aa-42ec-a945-5fd21dec0538",
		req: request{path: files.below(filesOf.path) + "/order"}, scopes: importScopes(),
		body: &payload{array: true, typed: valueBody(listValue("id",
			"the `ID` of a waiting file, the first to load first (repeatable): every waiting file").require())},
		conflict: fileOrderConflict(list),
	}
}

// fileDelete removes a waiting file from files, the file queue of a stream of
// filesOf (nested), addressed as one item of it (childItemTarget).
func fileDelete(filesOf resource[importStream], files child[queuedFile[importFile]]) *write {
	remove := childItemTarget(filesOf, files, deleteUse, nil)
	w := deleteWrite(remove.use, "Remove a waiting file from a stream's queue",
		"Removes the waiting file ID from the queue of the stream "+filesOf.argument()+" and deletes its\n"+
			"content. Only a waiting file is removed: one that a run claimed or imported is a conflict\n"+
			"(exit 5), as is any while the stream's job has a run in flight."+childItemHelp(filesOf, files, &remove),
		request{path: remove.path}, importScopes())
	w.conflict = fileDeleteConflict(queueList(filesOf, files))
	return w
}
