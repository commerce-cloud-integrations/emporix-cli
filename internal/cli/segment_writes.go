package cli

import (
	"net/http"
	"slices"

	"github.com/spf13/cobra"
)

// segmentManageScope is the scope of every write of the Customer Segment
// Service.
const segmentManageScope = "customersegment.segment_manage"

// maxSegmentBulk is the most items that a bulk write of the Customer Segment
// Service takes in one request.
const maxSegmentBulk = 200

// withAssignments is the switch of a delete of segments that deletes their
// assignments too (forceDelete).
func withAssignments() []paramFlag {
	return []paramFlag{{param: forceDeleteParam, flag: "with-assignments",
		usage: "delete the item and customer assignments of a segment with it (its IAM group assignments " +
			"go either way)"}}
}

// segmentPatchLanguageHelp says, in the help of segments patch, how its
// operations give localized text: as every write of mapsOnly, but for the
// value of a path to one translation.
const segmentPatchLanguageHelp = mapsOnlyHelp + "\nIn an operation, a path to one translation, such as /name/de, " +
	"takes a string in that\nlanguage."

// withAssignmentsNote follows what a delete of segments sends, in its help.
const withAssignmentsNote = "Emporix refuses to delete a segment that has items or customers assigned to it\n" +
	"(400 Bad Request, exit 1) unless --with-assignments deletes those assignments with it. It\n" +
	"deletes the IAM group assignments of a segment with it either way: a segment with only\n" +
	"IAM groups assigned is deleted without --with-assignments."

// segmentWrites returns the writes of segments: create, update (an upsert),
// patch (JSON Patch operations) and delete, and bulk-create, bulk-update and
// bulk-delete. Each but the deletes sends localized text, such as the name,
// as maps of translations only (mapsOnly): Emporix refuses a string in any
// language, and a map sent with one language may hold that language only.
func segmentWrites(a *app, segments resource[segment]) []*cobra.Command {
	scopes := []string{segmentManageScope}
	item, ids := segments.itemPath(), segments.itemIDs()
	id := segments.idSource() + "."
	bulk := &bulkWrites{
		path: segments.path, nouns: segments.nouns(), scopes: scopes, maxItems: maxSegmentBulk, language: mapsOnly,
		create: "Creates the segments in --file, a JSON array of segments, each as create takes it.",
		update: "Replaces the segments in --file, a JSON array of segments, each with its id, or creates\n" +
			"those that do not exist (an upsert).",
		delete: "Deletes the segments whose IDs --file gives, a JSON array such as [\"s1\", \"s2\"].\n" +
			withAssignmentsNote + unknownAsDeleted("ID"),
		deleteSwitches: withAssignments(),
	}
	return append([]*cobra.Command{
		(&write{
			use: createUse, short: "Create a segment", method: http.MethodPost, language: mapsOnly,
			long: "Creates the segment in --file, a JSON object such as `epx segments get ID -o json`\n" +
				"prints one, with its name, siteCode, validity and status (INACTIVE unless it says\n" +
				"ACTIVE); Emporix makes its id unless the body gives one. A segment with that id that\n" +
				"exists already is a conflict (exit 5).",
			req: request{path: segments.path}, scopes: scopes, body: &payload{},
		}).command(a),
		(&write{
			use: segments.itemUse("update"), short: "Upsert a segment", method: http.MethodPut, language: mapsOnly,
			long: "Replaces the segment ID with the one in --file, a JSON object as\n" +
				"`epx segments get ID -o json` prints it, or creates it with that id where none exists\n" +
				"(an upsert); members it leaves out, such as a description, are removed.\n" + id + "\n" +
				ifUnchangedNote,
			req: request{path: item}, scopes: scopes, body: &payload{}, reserved: bulk.reserved(http.MethodPut, item),
			ids: ids,
		}).command(a),
		(&write{
			use: segments.itemUse("patch"), short: "Change a segment", method: http.MethodPatch, language: mapsOnly,
			long: "Changes the segment ID with the JSON Patch operations (RFC 6902) of --file, a JSON array\n" +
				"such as [{\"op\": \"replace\", \"path\": \"/status\", \"value\": \"INACTIVE\"}]; epx refuses\n" +
				"anything else, which Emporix would answer as done without applying it, although it\n" +
				"raises the segment's metadata.version.\n" + id,
			languageHelp: segmentPatchLanguageHelp,
			req:          request{path: item}, scopes: scopes, body: operationsFile(), ids: ids,
		}).command(a),
		(&write{
			use: segments.itemUse("delete"), short: "Delete a segment", method: http.MethodDelete,
			long: "Deletes the segment ID. " + id + "\n" + withAssignmentsNote + unknownAsDeleted("ID"),
			req:  request{path: item}, scopes: scopes, body: noBody{}, switches: withAssignments(),
			reserved: bulk.reserved(http.MethodDelete, item), ids: ids,
			deletes: deletion{many: "the segment and its item, customer and IAM group assignments", with: "with-assignments"},
		}).command(a),
	}, bulk.commands(a)...)
}

// assignmentNote follows what a write that assigns something to a segment
// sends, in its help.
const assignmentNote = "--file may give the assignment's mixins and metadata (its version and the schemas of\n" +
	"its mixins); without it, epx sends an empty object."

// segmentAssignment declares one kind of item assigned to a segment, such as a
// customer without a legal entity, as segmentAssignmentWrites writes it.
type segmentAssignment struct {
	suffix  string              // of the names of its writes, e.g. "-b2b" in update-b2b; none when empty
	columns []string            // the columns of the child's table whose values name one; its first when empty
	noun    string              // one of them in summaries, e.g. "an IAM group"
	what    string              // one of them in help, by its arguments, e.g. "the IAM group GROUP"
	note    string              // what update's help adds after about, from a line of its own, e.g. what Emporix refuses
	about   string              // where its arguments come from, a sentence
	choices map[string][]string // the only values of its arguments, by name; any when nil
}

// segmentAssignmentWrites returns the writes of kind, items of c below a segment of
// segments: update, which assigns one (a PUT that creates the assignment or
// replaces it, an upsert), whose body --file gives, else an empty object,
// and delete, which removes it, each named with kind's suffix and addressing
// its item as a read of it does (childItemTarget). Where its ID would address
// a bulk write of bulk, the bulk writes of c, if any, they refuse it
// (bulkWrites.reserved).
func segmentAssignmentWrites[T any](a *app, segments resource[segment], c child[T], bulk *bulkWrites,
	kind *segmentAssignment) []*cobra.Command {
	assign := childItemTarget(segments, c, "update"+kind.suffix, kind.columns)
	remove := childItemTarget(segments, c, "delete"+kind.suffix, kind.columns)
	scopes, segment := []string{segmentManageScope}, segments.argument()
	// update completes the segment only, as it may assign any item of its kind;
	// delete completes the items assigned too.
	ofSegment := argumentIDs{segment: segments.ids()}
	return []*cobra.Command{
		(&write{
			use: assign.use, short: "Assign " + kind.noun + " to a segment", method: http.MethodPut,
			long: "Assigns " + kind.what + " to the segment\n" + segment + ", or replaces its assignment (an upsert).\n" +
				kind.about + kind.note + "\n" + assignmentNote + "\n" + ifUnchangedNote,
			req: request{path: assign.path}, scopes: scopes, body: emptyObjectOrFile(), choices: kind.choices,
			reserved: bulk.reserved(http.MethodPut, assign.path), ids: ofSegment,
		}).command(a),
		(&write{
			use: remove.use, short: "Remove " + kind.noun + " from a segment", method: http.MethodDelete,
			long: "Removes " + kind.what + " from the segment\n" + segment + "; Emporix answers one that is not\n" +
				"assigned as removed.\n" + kind.about,
			req: request{path: remove.path}, scopes: scopes, body: noBody{}, choices: kind.choices,
			reserved: bulk.reserved(http.MethodDelete, remove.path), ids: childItemIDs(segments, c, &remove),
		}).command(a),
	}
}

// oneCustomerKindNote is what Emporix refuses to assign to a segment of
// customers, in the help of the writes that assign them.
const oneCustomerKindNote = "\nA segment holds customers without a legal entity (B2C) or with one (B2B), never " +
	"both:\nEmporix refuses to assign one of the other kind."

// segmentCustomerWrites returns the writes of the customers of a segment of
// segments: update and delete, of a customer without (B2C) and with a legal
// entity (B2B: update-b2b, delete-b2b), and both in bulk, for either.
func segmentCustomerWrites(a *app, segments resource[segment], customers child[segmentCustomer]) []*cobra.Command {
	segment := segments.argument()
	bulk := &bulkWrites{
		path: customers.below(segments.path), args: segment, nouns: "customer assignments of a segment",
		scopes: []string{segmentManageScope}, maxItems: maxSegmentBulk, ids: argumentIDs{segment: segments.ids()},
		update: "Assigns the customers in --file to the segment " + segment + ", or replaces their assignments\n" +
			"(an upsert): a JSON array of assignments, each with its customerId, for B2B its\n" +
			"legalEntityId, and optionally its mixins and metadata." + oneCustomerKindNote,
		delete: "Removes the customers in --file from the segment " + segment + ", a JSON array such as\n" +
			"[{\"customerId\": \"c1\", \"legalEntityId\": \"le1\"}] (without legalEntityId for B2C).",
	}
	customer := "CUSTOMER the customer's id (the id of `epx customers get NUMBER -o json`), not its\nNUMBER"
	return slices.Concat(
		segmentAssignmentWrites(a, segments, customers, bulk, &segmentAssignment{
			noun: "a customer", what: "the customer CUSTOMER without a legal entity (B2C)",
			note:  oneCustomerKindNote,
			about: segments.argumentSource() + ", and\n" + customer + ".",
		}),
		segmentAssignmentWrites(a, segments, customers, bulk, &segmentAssignment{
			suffix: "-b2b", columns: []string{columnCustomer, columnLegalEntity}, noun: "a customer of a legal entity",
			what: "the customer CUSTOMER of the legal entity LEGAL_ENTITY (B2B)",
			note: oneCustomerKindNote,
			about: segments.argumentSource() + ",\n" + customer + ", and LEGAL_ENTITY the ID that\n" +
				"`epx legal-entities list` shows.",
		}),
		bulk.commands(a),
	)
}

// segmentGroupWrites returns the writes of the IAM groups of a segment of
// segments: update and delete.
func segmentGroupWrites(a *app, segments resource[segment], groups child[segmentGroup]) []*cobra.Command {
	return segmentAssignmentWrites(a, segments, groups, nil, &segmentAssignment{
		noun: "an IAM group", what: "the IAM group GROUP, a group of customers (userType CUSTOMER),",
		about: segments.argumentSource() + ", and GROUP the ID of an IAM group\nthat `epx iam groups list` shows.",
	})
}

// unknownItemNote is what Emporix refuses to assign to a segment as its item,
// in the help of the writes that assign them.
const unknownItemNote = "\nEmporix refuses an ID of no product or category of TYPE."

// segmentItemWrites returns the writes of the items of a segment of
// segments, products or categories: update and delete, and both in bulk,
// each of one TYPE.
func segmentItemWrites(a *app, segments resource[segment], items child[segmentItem]) []*cobra.Command {
	types := segments.argumentSource() + ",\nand " + segmentItemTypeSentence() + "."
	bulk := &bulkWrites{
		path: items.below(segments.path) + "/" + placeholderInner, args: segments.argument() + " " + columnType,
		nouns: "item assignments of a segment", scopes: []string{segmentManageScope}, maxItems: maxSegmentBulk,
		choices: segmentItemTypes(), ids: argumentIDs{segments.argument(): segments.ids()},
		update: "Assigns the products or categories (TYPE) in --file to the segment SEGMENT_ID, or replaces\n" +
			"their assignments (an upsert): a JSON array of assignments, each with the id of its item\n" +
			"and optionally its mixins and metadata.\n" + types + unknownItemNote,
		delete: "Removes the products or categories (TYPE) whose ids --file gives from the segment\n" +
			"SEGMENT_ID, a JSON array such as [\"100200\", \"100300\"].\n" + types,
	}
	return append(segmentAssignmentWrites(a, segments, items, bulk, &segmentAssignment{
		columns: []string{columnType, columnID}, noun: "a product or category", what: "the product or category ID",
		about: segments.argumentSource() + ", " + segmentItemTypeSentence() + ", and\n" +
			"ID the id of the product or category (as `epx products list` or `epx categories list`\nshows it).",
		note: unknownItemNote, choices: segmentItemTypes(),
	}), bulk.commands(a)...)
}
