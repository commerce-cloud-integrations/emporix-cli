package cli

import "github.com/spf13/cobra"

const versionedWritesNote = "create and update ID (an upsert) send the JSON body of --file; delete ID sends none."

type versioned struct {
	members    string
	required   string
	createNote string
	updateNote string
	deleteNote string
	unknown    unknownParent
	deletes    string
}

func (r resource[T]) versionedWrites(a *app, scopes []string, v *versioned) []*cobra.Command {
	del := r.deleteWrite(scopes, joinSentences(r.versionedDeleteHelp(v.unknown), v.deleteNote))
	del.deletes = deletion{many: v.deletes}
	return []*cobra.Command{
		r.createWrite(scopes, joinSentences(r.versionedCreateHelp(v.members), v.createNote)).command(a),
		r.upsertWrite(scopes, joinSentences(r.versionedUpsertHelp(v.required), v.updateNote)).command(a),
		del.command(a),
	}
}

func (r resource[T]) versionedCreateHelp(members string) string {
	return "Creates the " + r.singular + " that the JSON body of --file describes, shaped as " + r.getJSON() +
		" prints one: " + members + " Without an id, Emporix makes one, and answers with it; an id that exists " +
		"already is a conflict (exit 5)."
}

func (r resource[T]) versionedUpsertHelp(required string) string {
	return "Replaces the " + r.singular + " " + r.header[0] + " with the JSON body of --file, shaped as create " +
		"takes it but without an id, or creates it with that id where none exists (an upsert); it requires " +
		required + ". With metadata.version as " + r.getJSON() + " prints it, Emporix replaces it only if " +
		"nobody changed it since, and answers 409 Conflict (exit 5) otherwise; without metadata.version, it " +
		"replaces it whatever its version. " + r.idSource() + "."
}

func (r resource[T]) versionedDeleteHelp(unknown unknownParent) string {
	return "Deletes the " + r.singular + " " + r.header[0] + "; it sends no body. " + r.idSource() +
		"; an unknown one " + unknown.String() + "."
}

func (r resource[T]) getJSON() string {
	return "`epx " + r.getLine() + " -o json`"
}
