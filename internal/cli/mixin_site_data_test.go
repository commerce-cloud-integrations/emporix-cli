package cli

import "strings"

const (
	unmergedProduct = `{"code":"100200","id":"100200",` +
		`"metadata":{"mixins":{"marketAttributes":"https://example.com/m"}},` +
		`"mixins":{"marketAttributes":{"shippingRestricted":false,"taxCategory":"STANDARD"},` +
		`"marketAttributes__AcmeStore_NL":{"taxCategory":"REDUCED"}},"name":{"en":"Valve","nl":"Klep"},` +
		`"productType":"BASIC","published":true}`
	mergedProduct = `{"code":"100200","id":"100200",` +
		`"metadata":{"mixins":{"marketAttributes":"https://example.com/m"}},` +
		`"mixins":{"marketAttributes":{"shippingRestricted":false,"taxCategory":"REDUCED"}},` +
		`"name":{"en":"Valve","nl":"Klep"},"productType":"BASIC","published":true}`
	unmergedInstance = `{"id":"CAR1","metadata":{"mixins":{"carAttributes":"https://example.com/c"},"version":2},` +
		`"mixins":{"carAttributes":{"price":13},"carAttributes__AcmeStore_NL":{"price":99}},` +
		`"name":{"en":"Car","nl":"Auto"},"owner":{"type":"EMPLOYEE","userId":"u1"},"type":"CAR"}`
	mergedInstance = `{"id":"CAR1","metadata":{"mixins":{"carAttributes":"https://example.com/c"},"version":2},` +
		`"mixins":{"carAttributes":{"price":99}},"name":{"en":"Car","nl":"Auto"},` +
		`"owner":{"type":"EMPLOYEE","userId":"u1"},"type":"CAR"}`
)

const (
	unmergedProductYAML = `"code": "100200"
"id": "100200"
"metadata":
  "mixins":
    "marketAttributes": "https://example.com/m"
"mixins":
  "marketAttributes":
    "shippingRestricted": false
    "taxCategory": "STANDARD"
  "marketAttributes__AcmeStore_NL":
    "taxCategory": "REDUCED"
"name":
  "en": "Valve"
  "nl": "Klep"
"productType": "BASIC"
"published": true
`
	mergedProductYAML = `"code": "100200"
"id": "100200"
"metadata":
  "mixins":
    "marketAttributes": "https://example.com/m"
"mixins":
  "marketAttributes":
    "shippingRestricted": false
    "taxCategory": "REDUCED"
"name":
  "en": "Valve"
  "nl": "Klep"
"productType": "BASIC"
"published": true
`
	unmergedInstanceYAML = `"id": "CAR1"
"metadata":
  "mixins":
    "carAttributes": "https://example.com/c"
  "version": 2
"mixins":
  "carAttributes":
    "price": 13
  "carAttributes__AcmeStore_NL":
    "price": 99
"name":
  "en": "Car"
  "nl": "Auto"
"owner":
  "type": "EMPLOYEE"
  "userId": "u1"
"type": "CAR"
`
	mergedInstanceYAML = `"id": "CAR1"
"metadata":
  "mixins":
    "carAttributes": "https://example.com/c"
  "version": 2
"mixins":
  "carAttributes":
    "price": 99
"name":
  "en": "Car"
  "nl": "Auto"
"owner":
  "type": "EMPLOYEE"
  "userId": "u1"
"type": "CAR"
`
)

type siteAwareItem struct {
	unmerged, merged         string
	unmergedYAML, mergedYAML string
	tables, csv              map[string]string
}

func siteAwareProduct() siteAwareItem {
	head, row := "ID      CODE    TYPE   PUBLISHED  NAME\n", "100200  100200  BASIC  true       "
	csvHead, csvRow := "ID,CODE,TYPE,PUBLISHED,NAME\n", "100200,100200,BASIC,true,"
	return siteAwareItem{
		unmerged: unmergedProduct, merged: mergedProduct,
		unmergedYAML: unmergedProductYAML, mergedYAML: mergedProductYAML,
		tables: map[string]string{"en": head + row + "Valve\n", "nl": head + row + "Klep\n"},
		csv:    map[string]string{"en": csvHead + csvRow + "Valve\n", "nl": csvHead + csvRow + "Klep\n"},
	}
}

func siteAwareInstance() siteAwareItem {
	head, csvHead := "ID    NAME  OWNER        VERSION\n", "ID,NAME,OWNER,VERSION\n"
	return siteAwareItem{
		unmerged: unmergedInstance, merged: mergedInstance,
		unmergedYAML: unmergedInstanceYAML, mergedYAML: mergedInstanceYAML,
		tables: map[string]string{"en": head + "CAR1  Car   EMPLOYEE u1  2\n", "nl": head + "CAR1  Auto  EMPLOYEE u1  2\n"},
		csv:    map[string]string{"en": csvHead + "CAR1,Car,EMPLOYEE u1,2\n", "nl": csvHead + "CAR1,Auto,EMPLOYEE u1,2\n"},
	}
}

func (i *siteAwareItem) item(merged bool) string {
	if merged {
		return i.merged
	}
	return i.unmerged
}

func (i *siteAwareItem) yaml(merged, listed bool) string {
	text := i.unmergedYAML
	if merged {
		text = i.mergedYAML
	}
	if !listed {
		return text
	}
	lines := strings.SplitAfter(strings.TrimSuffix(text, "\n"), "\n")
	for n := range lines {
		lines[n] = "  " + lines[n]
	}
	return "- " + strings.TrimPrefix(strings.Join(lines, ""), "  ") + "\n"
}
