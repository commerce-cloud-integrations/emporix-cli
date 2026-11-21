package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

type formPart struct {
	flag  string
	bind  func(cmd *cobra.Command) addPart
	bound int
}

type addPart func(ctx context.Context, parts []emporix.Part) ([]emporix.Part, error)

type form []formPart

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
	}, "a multipart/form-data body, which the flags make."
}

func filePart(name, flag, usage string) formPart {
	return filePartOf(name, bodyFile{flag: flag, limit: maxWriteFileBytes}, usage, jsonContent).require()
}

func (p formPart) require() formPart {
	bind := p.bind
	p.bind = func(cmd *cobra.Command) addPart {
		add := bind(cmd)
		requireFlag(cmd, p.flag)
		return add
	}
	return p
}

func jsonContent(_ string, content []byte) (string, error) {
	if err := isJSON(content); err != nil {
		return "", err
	}
	return emporix.JSONContentType, nil
}

func filePartOf(name string, file bodyFile, usage string,
	typeOf func(fileName string, content []byte) (string, error)) formPart {
	return formPart{flag: file.flag, bind: func(cmd *cobra.Command) addPart {
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

func fieldPart(name, flag, usage string) formPart {
	return formPart{flag: flag, bind: func(cmd *cobra.Command) addPart {
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

type formFile struct {
	form        form
	field       string
	usage       string
	refuseAlone func(body json.RawMessage) error
	parts       func(ctx context.Context) ([]emporix.Part, error)
	file        *givenBody
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

type formBody struct {
	form  form
	oneOf []string
	parts func(ctx context.Context) ([]emporix.Part, error)
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

func encodeForm(parts []emporix.Part, limit int) (emporix.Body, error) {
	form, err := emporix.FormBody(parts)
	if err != nil || limit == 0 || form.Len() <= limit {
		return form, err
	}
	return emporix.Body{}, usageError{fmt.Errorf("the form is %d bytes, larger than the %d bytes that Emporix "+
		"takes in one request: send a smaller file or body", form.Len(), limit)}
}

func (*formBody) fromStdin() bool { return false }
