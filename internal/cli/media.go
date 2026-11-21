package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

type mediaAsset struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Access  string `json:"access"`
	Status  string `json:"status"`
	Details struct {
		Filename string      `json:"filename"`
		MimeType string      `json:"mimeType"`
		Bytes    json.Number `json:"bytes"`
	} `json:"details"`
	RefIDs []assetReference `json:"refIds"`
}

type assetReference struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (r assetReference) String() string { return spaced(r.Type, r.ID) }

type downloadURL struct {
	URL       string `json:"url"`
	Provider  string `json:"provider"`
	ExpiresAt string `json:"expiresAt"`
}

func newMediaCommand(a *app) *cobra.Command {
	assets := resource[mediaAsset]{
		use:   "media",
		noun:  "media assets",
		short: "Read, upload, link and change media assets",
		long: "Reads media assets: files Emporix stores (BLOB) or links to files elsewhere (LINK). ACCESS " +
			"is PUBLIC or PRIVATE; STATUS is PENDING until an upload completes, then READY or empty; " +
			"BYTES is the file size as sent, and REFERENCES the items an asset belongs to (type and " +
			"id). URLs are only in `-o json`; `epx media download-url` gives one that downloads the file.",
		path:    "/media/{tenant}/assets",
		scopes:  []string{"media.asset_read"},
		filters: listFilters{kinds: filterable},
		view: view[mediaAsset]{
			singular: "asset",
			header:   []string{columnID, columnType, "ACCESS", columnStatus, "FILE", "MIME TYPE", "BYTES", "REFERENCES"},
			row: func(m mediaAsset, _ language) []string {
				return []string{
					m.ID, m.Type, m.Access, m.Status, m.Details.Filename, m.Details.MimeType, m.Details.Bytes.String(),
					joinEach(m.RefIDs, ", ", assetReference.String),
				}
			},
		},
	}
	return newResourceCommand(a, assets, append([]*cobra.Command{childCommand(a, assets, child[downloadURL]{
		use:   "download-url",
		short: "Show a URL that downloads an asset",
		long: "Reads a URL that downloads an asset. For a private stored file it is a signed URL that " +
			"works without a token until EXPIRES (15 minutes by default), so treat it as a secret; for " +
			"a public file or a link, PROVIDER CLOUDINARY or LINK, it is the asset's permanent URL. " +
			"Reading the URL of a pending upload whose file is stored makes Emporix complete it; one " +
			"whose file is not stored yet fails.",
		fetch: single{items: oneItem},
		view: view[downloadURL]{
			singular: "download URL",
			header:   []string{columnProvider, "EXPIRES", columnURL},
			row:      func(d downloadURL, _ language) []string { return []string{d.Provider, d.ExpiresAt, d.URL} },
		},
	})}, mediaWrites(a, assets)...)...)
}
