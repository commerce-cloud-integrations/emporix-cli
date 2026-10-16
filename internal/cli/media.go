package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

// mediaAsset is the part of a Media Service asset shown in tables.
type mediaAsset struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Access  string `json:"access"`
	Status  string `json:"status"`
	Details struct {
		Filename string      `json:"filename"`
		MimeType string      `json:"mimeType"`
		Bytes    json.Number `json:"bytes"` // kept as sent
	} `json:"details"`
	RefIDs []assetReference `json:"refIds"`
}

// assetReference is an item an asset belongs to, such as a product.
type assetReference struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (r assetReference) String() string { return spaced(r.Type, r.ID) }

// downloadURL is a URL that downloads an asset.
type downloadURL struct {
	URL       string `json:"url"`
	Provider  string `json:"provider"`
	ExpiresAt string `json:"expiresAt"`
}

// newMediaCommand reads media assets and a URL to download one, and creates,
// uploads and changes them (mediaWrites).
func newMediaCommand(a *app) *cobra.Command {
	assets := resource[mediaAsset]{
		use:   "media",
		noun:  "media assets",
		short: "Read, upload, link and change media assets",
		long: "Reads media assets: files Emporix stores (BLOB) or links to files elsewhere (LINK). ACCESS\n" +
			"is PUBLIC or PRIVATE; STATUS is PENDING until an upload completes, then READY or empty;\n" +
			"BYTES is the file size as sent, and REFERENCES the items an asset belongs to (type and\n" +
			"id). URLs are only in -o json; `epx media download-url` gives one that downloads the file.",
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
		long: "Reads a URL that downloads an asset. For a private stored file it is a signed URL that\n" +
			"works without a token until EXPIRES (15 minutes by default), so treat it as a secret; for\n" +
			"a public file or a link, PROVIDER CLOUDINARY or LINK, it is the asset's permanent URL.\n" +
			"Reading the URL of a pending upload whose file is stored makes Emporix complete it; one\n" +
			"whose file is not stored yet fails.",
		fetch: single{items: oneItem},
		view: view[downloadURL]{
			singular: "download URL",
			header:   []string{columnProvider, "EXPIRES", columnURL},
			row:      func(d downloadURL, _ language) []string { return []string{d.Provider, d.ExpiresAt, d.URL} },
		},
	})}, mediaWrites(a, assets)...)...)
}
