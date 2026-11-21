package cli

import (
	"net/http"
	"slices"

	"github.com/spf13/cobra"
)

const segmentManageScope = "customersegment.segment_manage"

const maxSegmentBulk = 200

func withAssignments() []paramFlag {
	return []paramFlag{{param: forceDeleteParam, flag: "with-assignments",
		usage: "delete the item and customer assignments of a segment with it (its IAM group assignments " +
			"go either way)"}}
}

func segmentPatchLanguageHelp() string {
	return mapsOnlyHelp() + " In an operation, a path to one translation, such as /name/de, " +
		"takes a string in that language."
}

const withAssignmentsNote = "Emporix refuses to delete a segment that has items or customers assigned to it " +
	"(400 Bad Request, exit 1) unless --with-assignments deletes those assignments with it. It " +
	"deletes the IAM group assignments of a segment with it either way: a segment with only " +
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
		update: "Replaces the segments in --file, a JSON array of segments, each with its id, or creates " +
			"those that do not exist (an upsert).",
		delete: "Deletes the segments whose IDs --file gives, a JSON array such as " + literal(`["s1", "s2"]`) + ". " +
			withAssignmentsNote + " " + unknownAsDeleted("ID"),
		deleteSwitches: withAssignments(),
	}
	return append([]*cobra.Command{
		(&write{
			use: createUse, short: "Create a segment", method: http.MethodPost, language: mapsOnly,
			long: "Creates the segment in --file, a JSON object such as `epx segments get ID -o json` " +
				"prints one, with its name, siteCode, validity and status (INACTIVE unless it says " +
				"ACTIVE); Emporix makes its id unless the body gives one. A segment with that id that " +
				"exists already is a conflict (exit 5).",
			req: request{path: segments.path}, scopes: scopes, body: &payload{},
		}).command(a),
		(&write{
			use: segments.itemUse("update"), short: "Upsert a segment", method: http.MethodPut, language: mapsOnly,
			long: "Replaces the segment ID with the one in --file, a JSON object as " +
				"`epx segments get ID -o json` prints it, or creates it with that id where none exists " +
				"(an upsert); members it leaves out, such as a description, are removed. " + id + " " +
				ifUnchangedNote,
			req: request{path: item}, scopes: scopes, body: &payload{}, reserved: bulk.reserved(http.MethodPut, item),
			ids: ids,
		}).command(a),
		(&write{
			use: segments.itemUse("patch"), short: "Change a segment", method: http.MethodPatch, language: mapsOnly,
			long: "Changes the segment ID with the JSON Patch operations (RFC 6902) of --file, a JSON array " +
				"such as " + literal(`[{"op": "replace", "path": "/status", "value": "INACTIVE"}]`) + "; epx refuses " +
				"anything else, which Emporix would answer as done without applying it, although it " +
				"raises the segment's metadata.version. " + id,
			languageHelp: segmentPatchLanguageHelp(),
			req:          request{path: item}, scopes: scopes, body: operationsFile(), ids: ids,
		}).command(a),
		(&write{
			use: segments.itemUse("delete"), short: "Delete a segment", method: http.MethodDelete,
			long: "Deletes the segment ID. " + id + " " + withAssignmentsNote + " " + unknownAsDeleted("ID"),
			req:  request{path: item}, scopes: scopes, body: noBody{}, switches: withAssignments(),
			reserved: bulk.reserved(http.MethodDelete, item), ids: ids,
			deletes: deletion{many: "the segment and its item, customer and IAM group assignments", with: "with-assignments"},
		}).command(a),
	}, bulk.commands(a)...)
}

const assignmentNote = "--file may give the assignment's mixins and metadata (its version and the schemas of " +
	"its mixins); without it, epx sends an empty object."

type segmentAssignment struct {
	suffix      string
	columns     []string
	summaryNoun string
	helpNoun    string
	assignNote  string
	argSource   string
	choices     map[string][]string
}

func segmentAssignmentWrites[T any](a *app, segments resource[segment], c child[T], bulk *bulkWrites,
	kind *segmentAssignment) []*cobra.Command {
	assign := childItemTarget(segments, c, "update"+kind.suffix, kind.columns)
	remove := childItemTarget(segments, c, "delete"+kind.suffix, kind.columns)
	scopes, segment := []string{segmentManageScope}, segments.argument()
	ofSegment := argumentIDs{segment: segments.ids()}
	return []*cobra.Command{
		(&write{
			use: assign.use, short: "Assign " + kind.summaryNoun + " to a segment", method: http.MethodPut,
			long: joinSentences("Assigns "+kind.helpNoun+" to the segment "+segment+", or replaces its assignment (an upsert).",
				kind.argSource, kind.assignNote, assignmentNote, ifUnchangedNote),
			req: request{path: assign.path}, scopes: scopes, body: emptyObjectOrFile(), choices: kind.choices,
			reserved: bulk.reserved(http.MethodPut, assign.path), ids: ofSegment,
		}).command(a),
		(&write{
			use: remove.use, short: "Remove " + kind.summaryNoun + " from a segment", method: http.MethodDelete,
			long: "Removes " + kind.helpNoun + " from the segment " + segment + "; Emporix answers one that is not " +
				"assigned as removed. " + kind.argSource,
			req: request{path: remove.path}, scopes: scopes, body: noBody{}, choices: kind.choices,
			reserved: bulk.reserved(http.MethodDelete, remove.path), ids: childItemIDs(segments, c, &remove),
		}).command(a),
	}
}

const oneCustomerKindNote = "A segment holds customers without a legal entity (B2C) or with one (B2B), never " +
	"both: Emporix refuses to assign one of the other kind."

func segmentCustomerWrites(a *app, segments resource[segment], customers child[segmentCustomer]) []*cobra.Command {
	segment := segments.argument()
	bulk := &bulkWrites{
		path: customers.below(segments.path), args: segment, nouns: "customer assignments of a segment",
		scopes: []string{segmentManageScope}, maxItems: maxSegmentBulk, ids: argumentIDs{segment: segments.ids()},
		update: "Assigns the customers in --file to the segment " + segment + ", or replaces their assignments " +
			"(an upsert): a JSON array of assignments, each with its customerId, for B2B its " +
			"legalEntityId, and optionally its mixins and metadata. " + oneCustomerKindNote,
		delete: "Removes the customers in --file from the segment " + segment + ", a JSON array such as " +
			literal(`[{"customerId": "c1", "legalEntityId": "le1"}]`) + " (without legalEntityId for B2C).",
	}
	customer := "CUSTOMER the customer's id (the id of `epx customers get NUMBER -o json`), not its NUMBER"
	return slices.Concat(
		segmentAssignmentWrites(a, segments, customers, bulk, &segmentAssignment{
			summaryNoun: "a customer", helpNoun: "the customer CUSTOMER without a legal entity (B2C)",
			assignNote: oneCustomerKindNote,
			argSource:  segments.argumentSource() + ", and " + customer + ".",
		}),
		segmentAssignmentWrites(a, segments, customers, bulk, &segmentAssignment{
			suffix: "-b2b", columns: []string{columnCustomer, columnLegalEntity}, summaryNoun: "a customer of a legal entity",
			helpNoun:   "the customer CUSTOMER of the legal entity LEGAL_ENTITY (B2B)",
			assignNote: oneCustomerKindNote,
			argSource: segments.argumentSource() + ", " + customer + ", and LEGAL_ENTITY the ID that " +
				"`epx legal-entities list` shows.",
		}),
		bulk.commands(a),
	)
}

func segmentGroupWrites(a *app, segments resource[segment], groups child[segmentGroup]) []*cobra.Command {
	return segmentAssignmentWrites(a, segments, groups, nil, &segmentAssignment{
		summaryNoun: "an IAM group", helpNoun: "the IAM group GROUP, a group of customers (userType CUSTOMER),",
		argSource: segments.argumentSource() + ", and GROUP the ID of an IAM group that `epx iam groups list` shows.",
	})
}

const unknownItemNote = "Emporix refuses an ID of no product or category of TYPE."

func segmentItemWrites(a *app, segments resource[segment], items child[segmentItem]) []*cobra.Command {
	types := segments.argumentSource() + ", and " + segmentItemTypeSentence() + "."
	bulk := &bulkWrites{
		path: items.below(segments.path) + "/" + placeholderInner, args: segments.argument() + " " + columnType,
		nouns: "item assignments of a segment", scopes: []string{segmentManageScope}, maxItems: maxSegmentBulk,
		choices: segmentItemTypes(), ids: argumentIDs{segments.argument(): segments.ids()},
		update: "Assigns the products or categories (TYPE) in --file to the segment SEGMENT_ID, or replaces " +
			"their assignments (an upsert): a JSON array of assignments, each with the id of its item " +
			"and optionally its mixins and metadata. " + types + " " + unknownItemNote,
		delete: "Removes the products or categories (TYPE) whose ids --file gives from the segment " +
			"SEGMENT_ID, a JSON array such as " + literal(`["100200", "100300"]`) + ". " + types,
	}
	return append(segmentAssignmentWrites(a, segments, items, bulk, &segmentAssignment{
		columns: []string{columnType, columnID}, summaryNoun: "a product or category", helpNoun: "the product or category ID",
		argSource: segments.argumentSource() + ", " + segmentItemTypeSentence() +
			", and ID the id of the product or category (as `epx products list` or `epx categories list` shows it).",
		assignNote: unknownItemNote, choices: segmentItemTypes(),
	}), bulk.commands(a)...)
}
