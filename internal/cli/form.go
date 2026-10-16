package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// A form is a multipart/form-data body, such as a JSON schema file that the
// Schema Service parses, or stores with the JSON that describes it. Flags
// give its parts (formPart): the body of a read (form) or, followed by the
// JSON of --file, of a write (formFile). A dry run describes its parts and
// shows the content of the JSON ones only (output.WriteRequest).

// formPart is a part of a form that a flag gives (filePartOf, fieldPart).
type formPart struct {
	// bind adds the flag to cmd and returns how a run adds the part.
	bind func(cmd *cobra.Command) addPart
	// bound is the most bytes that Emporix takes in a request whose form
	// holds the part, such as maxUploadBytes for a file to upload (uploadPart);
	// 0 for any.
	bound int
}

// addPart adds to the parts of a form in a run the part that a flag gives.
type addPart func(ctx context.Context, parts []emporix.Part) ([]emporix.Part, error)

// form is the form whose parts flags give, in their order: the body of a read
// (readBody), and the start of the form of a write (formFile).
type form []formPart

// parts adds the flags of f's parts to cmd and returns how a run makes the
// parts, in their order.
func (f form) parts(cmd *cobra.Command) func(ctx context.Context) ([]emporix.Part, error) {
	adds := make([]addPart, len(f))
	for i, part := range f {
		adds[i] = part.bind(cmd)
	}
	return func(ctx context.Context) ([]emporix.Part, error) {
		var parts []emporix.Part
		for _, add := range adds {
			var err error
			if parts, err = add(ctx, parts); err != nil {
				return nil, err
			}
		}
		return parts, nil
	}
}

// limit returns the most bytes that Emporix takes in a request of f: the
// least bound of its parts, if any; 0 for any.
func (f form) limit() int {
	limit := 0
	for _, part := range f {
		if part.bound > 0 && (limit == 0 || part.bound < limit) {
			limit = part.bound
		}
	}
	return limit
}

func (f form) bindRead(cmd *cobra.Command) (made makeBody, sends string) {
	parts := f.parts(cmd)
	return func(ctx context.Context) (emporix.Body, error) {
		given, err := parts(ctx)
		if err != nil {
			return emporix.Body{}, err
		}
		return encodeForm(given, f.limit())
	}, "a multipart/form-data\nbody, which the flags make."
}

// filePart is the part name of a form that sends the JSON file that --flag
// (described by usage) names, which it requires, as filePartOf sends a file,
// within the size of --file, as application/json; one that is not JSON is
// refused (isJSON).
func filePart(name, flag, usage string) formPart {
	part := filePartOf(name, bodyFile{flag: flag, limit: maxWriteFileBytes}, usage, jsonContent)
	return formPart{bind: func(cmd *cobra.Command) addPart {
		add := part.bind(cmd)
		requireFlag(cmd, flag)
		return add
	}}
}

// jsonContent returns the content type of a JSON file: application/json, or,
// for content that is not JSON, why it cannot be sent (isJSON).
func jsonContent(_ string, content []byte) (string, error) {
	if err := isJSON(content); err != nil {
		return "", err
	}
	return emporix.JSONContentType, nil
}

// filePartOf is the part name of a form that sends the file that the flag of
// file (described by usage) names, when it is given: its content byte for
// byte, read as --file is (bodyFile) up to file.limit, with the name of the
// file, never its directory, and the content type that typeOf gives for the
// file's name and content, or why it cannot be sent (a usage error). It
// reads only from a path, as standard input has no name to send.
func filePartOf(name string, file bodyFile, usage string,
	typeOf func(fileName string, content []byte) (string, error)) formPart {
	return formPart{bind: func(cmd *cobra.Command) addPart {
		given := file.bind(cmd, usage)
		return func(ctx context.Context, parts []emporix.Part) ([]emporix.Part, error) {
			if !given.given() {
				return parts, nil
			}
			fileName, err := given.fileName()
			if err != nil {
				return nil, err
			}
			content, err := given.read(ctx)
			if err != nil {
				return nil, err
			}
			contentType, err := typeOf(fileName, content)
			if err != nil {
				return nil, given.failure(given.file, err)
			}
			return append(parts, emporix.Part{Name: name, FileName: fileName, ContentType: contentType, Content: content}),
				nil
		}
	}}
}

// fieldPart is the field name of a form that --flag (described by usage)
// gives, sent as given when the flag is given; an empty value is refused
// while flags are parsed (paramValue).
func fieldPart(name, flag, usage string) formPart {
	return formPart{bind: func(cmd *cobra.Command) addPart {
		value := &paramValue{paramFlag: paramFlag{param: name, flag: flag, usage: usage}}
		value.register(cmd)
		return func(_ context.Context, parts []emporix.Part) ([]emporix.Part, error) {
			if value.given == "" {
				return parts, nil
			}
			return append(parts, emporix.Part{Name: value.param, Content: []byte(value.given)}), nil
		}
	}}
}

// formFile is the body of a write that is a form (a bodySource): the parts
// of form, then the JSON of --file (described by usage), which it requires,
// sent byte for byte as the field that field names, e.g. body. Where no part
// of form is given, as a media asset without a file to upload, the JSON of
// --file alone is the body, unless refuseAlone refuses it. Each write takes its
// own.
type formFile struct {
	form  form
	field string
	usage string // the help of --file
	// refuseAlone returns why the JSON of --file cannot be the body alone,
	// without a part of form, as Emporix would refuse it; nil takes any.
	refuseAlone func(body json.RawMessage) error
	parts       func(ctx context.Context) ([]emporix.Part, error) // the parts of form in a run, once bound
	file        *givenBody                                        // --file, once bound
}

func (f *formFile) bind(cmd *cobra.Command) string {
	f.parts = f.form.parts(cmd)
	f.file = bodyFile{flag: fileFlag, limit: maxWriteFileBytes}.bind(cmd, f.usage+stdinFileNote)
	requireFlag(cmd, fileFlag)
	return "--" + fileFlag + " -"
}

func (*formFile) scopes() []string { return nil }

func (f *formFile) fill(ctx context.Context, w *emporix.Write, _ []string) error {
	given, err := f.parts(ctx)
	if err != nil {
		return err
	}
	content, err := f.file.readJSON(ctx)
	if err != nil {
		return err
	}
	if len(given) == 0 {
		return f.fillAlone(w, content)
	}
	w.Body, err = encodeForm(append(given,
		emporix.Part{Name: f.field, ContentType: emporix.JSONContentType, Content: content}), f.form.limit())
	return err
}

// fillAlone sets the body of w to body, the JSON of --file alone, unless
// refuseAlone refuses it: then a usage error naming the file.
func (f *formFile) fillAlone(w *emporix.Write, body json.RawMessage) (err error) {
	if f.refuseAlone != nil {
		if refused := f.refuseAlone(body); refused != nil {
			return f.file.failure(f.file.file, refused)
		}
	}
	w.Body, err = emporix.JSONBody(body)
	return err
}

func (f *formFile) fromStdin() bool { return f.file.fromStdin() }

// formBody is the body of a write that is a form whose parts flags give
// (form), exactly one of the flags oneOf, such as a file to upload or the ID
// of one that Emporix stores already. Each write takes its own.
type formBody struct {
	form  form
	oneOf []string
	parts func(ctx context.Context) ([]emporix.Part, error) // the parts of form in a run, once bound
}

func (f *formBody) bind(cmd *cobra.Command) string {
	f.parts = f.form.parts(cmd)
	cmd.MarkFlagsOneRequired(f.oneOf...)
	cmd.MarkFlagsMutuallyExclusive(f.oneOf...)
	return ""
}

func (*formBody) scopes() []string { return nil }

func (f *formBody) fill(ctx context.Context, w *emporix.Write, _ []string) error {
	given, err := f.parts(ctx)
	if err != nil {
		return err
	}
	w.Body, err = encodeForm(given, f.form.limit())
	return err
}

// encodeForm returns the form of parts (emporix.FormBody), unless limit is
// not 0 and the form is larger than limit bytes, the most that Emporix takes
// in one request, such as maxUploadBytes for the form of a file to upload:
// then a usage error, as nothing was sent.
func encodeForm(parts []emporix.Part, limit int) (emporix.Body, error) {
	form, err := emporix.FormBody(parts)
	if err != nil || limit == 0 || form.Len() <= limit {
		return form, err
	}
	return emporix.Body{}, usageError{fmt.Errorf("the form is %d bytes, larger than the %d bytes that Emporix "+
		"takes in one request: send a smaller file or body", form.Len(), limit)}
}

// fromStdin implements bodySource: no part of a form comes from standard
// input (givenBody.fileName).
func (*formBody) fromStdin() bool { return false }
