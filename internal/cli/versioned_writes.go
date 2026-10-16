package cli

import (
	"cmp"

	"github.com/spf13/cobra"
)

// The Customer Management and Vendor services write their items alike, with
// the same write scope for each collection: create makes the id of an item
// whose body gives none and refuses one that exists; update ID creates a
// missing item (an upsert), replaces one whatever its version where the body
// has no metadata.version, and refuses a stale one with 409 Conflict; and
// delete ID answers an unknown item with 404 Not Found or, where the
// collection declares it, with 204 No Content as for one that it deleted.
// The help below says so for the writes that declare them
// (resource_writes.go).

// versionedWritesNote ends the help of a group with these writes.
const versionedWritesNote = "create and update ID (an upsert) send the JSON body of --file; delete ID sends none."

// versioned holds the parts of one collection's versioned writes that
// differ: the members of create's body, what update requires, the notes that
// end the help of each write, what a delete of an unknown item gives, and
// what a delete deletes besides the item.
type versioned struct {
	members    string        // the members of create's body, e.g. "name (required), ..."
	required   string        // what update's body requires, e.g. "name and type"
	createNote string        // ends create's help, if not empty
	updateNote string        // ends update's help, if not empty
	deleteNote string        // ends delete's help, if not empty
	unknown    unknownParent // what a delete of an unknown item gives; notFound when empty
	cascade    string        // what delete deletes besides the item (deletion.many), if anything
}

// versionedWrites returns create, update ID and delete ID of r with scopes,
// whose help is that of versionedCreateHelp, versionedUpsertHelp and
// versionedDeleteHelp, as v says.
func (r resource[T]) versionedWrites(a *app, scopes []string, v *versioned) []*cobra.Command {
	del := r.deleteWrite(scopes, withNote(r.versionedDeleteHelp(cmp.Or(v.unknown, notFound)), v.deleteNote))
	del.deletes = deletion{many: v.cascade}
	return []*cobra.Command{
		r.createWrite(scopes, withNote(r.versionedCreateHelp(v.members), v.createNote)).command(a),
		r.upsertWrite(scopes, withNote(r.versionedUpsertHelp(v.required), v.updateNote)).command(a),
		del.command(a),
	}
}

// withNote returns help ended by the line note, if not empty.
func withNote(help, note string) string {
	if note == "" {
		return help
	}
	return help + "\n" + note
}

// versionedCreateHelp returns the help of create of r, whose body holds
// members, help text of its own lines that ends without a line break.
func (r resource[T]) versionedCreateHelp(members string) string {
	return "Creates the " + r.singular + " that the JSON body of --file describes, shaped as\n" + r.getJSON() +
		" prints one:\n" + members + "\nWithout an id, Emporix makes one, and answers with it; an id that exists " +
		"already is a\nconflict (exit 5)."
}

// versionedUpsertHelp returns the help of update ID of r, whose body requires
// required, e.g. "name and type".
func (r resource[T]) versionedUpsertHelp(required string) string {
	return "Replaces the " + r.singular + " " + r.header[0] + " with the JSON body of --file, shaped as create " +
		"takes it\nbut without an id, or creates it with that id where none exists (an upsert); it requires\n" +
		required + ".\nWith metadata.version as " + r.getJSON() + " prints it, Emporix\nreplaces it only if " +
		"nobody changed it since, and answers 409 Conflict (exit 5) otherwise;\nwithout metadata.version, it " +
		"replaces it whatever its version.\n" + r.idSource() + "."
}

// versionedDeleteHelp returns the help of delete ID of r, of which an
// unknown item gives unknown.
func (r resource[T]) versionedDeleteHelp(unknown unknownParent) string {
	return "Deletes the " + r.singular + " " + r.header[0] + "; it sends no body.\n" + r.idSource() +
		"; an unknown one " + string(unknown) + "."
}

// getJSON returns the command line that prints one item of r as JSON, e.g.
// "`epx vendors get ID -o json`".
func (r resource[T]) getJSON() string {
	return "`epx " + r.commandLine(r.itemUse("get"), r.itemRequest()) + " -o json`"
}
