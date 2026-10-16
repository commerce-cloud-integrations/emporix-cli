# Customer Segment Service

> Which Customer Segment Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/customer-segments/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/customer-segments/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`de786a4`](https://github.com/emporix/api-references/commit/de786a4))
- Version: not stated in the reference file
- Operations: 36 (34 built, 2 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `POST /customer-segment/{tenant}/segments` | Built | [`epx segments create`](/emporix-cli/commands/segments#epx-segments-create) |
| `GET /customer-segment/{tenant}/segments` | Built | [`epx segments list`](/emporix-cli/commands/segments#epx-segments-list) |
| `POST /customer-segment/{tenant}/segments/search` | Built | [`epx segments search`](/emporix-cli/commands/segments#epx-segments-search) |
| `GET /customer-segment/{tenant}/segments/me` | Not applicable | Documented for a signed-in customer's token (customersegment.segment_read_own); a service token gets an empty answer. |
| `GET /customer-segment/{tenant}/segments/{segmentId}` | Built | [`epx segments get`](/emporix-cli/commands/segments#epx-segments-get) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}` | Built | [`epx segments update`](/emporix-cli/commands/segments#epx-segments-update) |
| `PATCH /customer-segment/{tenant}/segments/{segmentId}` | Built | [`epx segments patch`](/emporix-cli/commands/segments#epx-segments-patch) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}` | Built | [`epx segments delete`](/emporix-cli/commands/segments#epx-segments-delete) |
| `POST /customer-segment/{tenant}/segments/bulk` | Built | [`epx segments bulk-create`](/emporix-cli/commands/segments#epx-segments-bulk-create) |
| `PUT /customer-segment/{tenant}/segments/bulk` | Built | [`epx segments bulk-update`](/emporix-cli/commands/segments#epx-segments-bulk-update) |
| `DELETE /customer-segment/{tenant}/segments/bulk` | Built | [`epx segments bulk-delete`](/emporix-cli/commands/segments#epx-segments-bulk-delete) |
| `POST /customer-segment/{tenant}/segments/match` | Built | [`epx segments match`](/emporix-cli/commands/segments#epx-segments-match) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/customers` | Built | [`epx segments customers list`](/emporix-cli/commands/segments#epx-segments-customers-list) |
| `POST /customer-segment/{tenant}/segments/{segmentId}/customers/search` | Built | [`epx segments customers search`](/emporix-cli/commands/segments#epx-segments-customers-search) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}` | Built | [`epx segments customers get`](/emporix-cli/commands/segments#epx-segments-customers-get) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}` | Built | [`epx segments customers update`](/emporix-cli/commands/segments#epx-segments-customers-update) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}` | Built | [`epx segments customers delete`](/emporix-cli/commands/segments#epx-segments-customers-delete) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}/{legalEntityId}` | Built | [`epx segments customers get-b2b`](/emporix-cli/commands/segments#epx-segments-customers-get-b2b) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}/{legalEntityId}` | Built | [`epx segments customers update-b2b`](/emporix-cli/commands/segments#epx-segments-customers-update-b2b) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/customers/{customerId}/{legalEntityId}` | Built | [`epx segments customers delete-b2b`](/emporix-cli/commands/segments#epx-segments-customers-delete-b2b) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/customers/bulk` | Built | [`epx segments customers bulk-update`](/emporix-cli/commands/segments#epx-segments-customers-bulk-update) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/customers/bulk` | Built | [`epx segments customers bulk-delete`](/emporix-cli/commands/segments#epx-segments-customers-bulk-delete) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/groups` | Built | [`epx segments groups list`](/emporix-cli/commands/segments#epx-segments-groups-list) |
| `POST /customer-segment/{tenant}/segments/{segmentId}/groups/search` | Built | [`epx segments groups search`](/emporix-cli/commands/segments#epx-segments-groups-search) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/groups/{groupId}` | Built | [`epx segments groups get`](/emporix-cli/commands/segments#epx-segments-groups-get) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/groups/{groupId}` | Built | [`epx segments groups update`](/emporix-cli/commands/segments#epx-segments-groups-update) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/groups/{groupId}` | Built | [`epx segments groups delete`](/emporix-cli/commands/segments#epx-segments-groups-delete) |
| `GET /customer-segment/{tenant}/segments/items/category-trees` | Not applicable | Documented for a signed-in customer's token (customersegment.segment_read_own); a service token gets an empty answer. |
| `GET /customer-segment/{tenant}/segments/items` | Built | [`epx segments all-items`](/emporix-cli/commands/segments#epx-segments-all-items) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/items` | Built | [`epx segments items list`](/emporix-cli/commands/segments#epx-segments-items-list) |
| `POST /customer-segment/{tenant}/segments/{segmentId}/items/search` | Built | [`epx segments items search`](/emporix-cli/commands/segments#epx-segments-items-search) |
| `GET /customer-segment/{tenant}/segments/{segmentId}/items/{type}/{itemId}` | Built | [`epx segments items get`](/emporix-cli/commands/segments#epx-segments-items-get) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/items/{type}/{itemId}` | Built | [`epx segments items update`](/emporix-cli/commands/segments#epx-segments-items-update) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/items/{type}/{itemId}` | Built | [`epx segments items delete`](/emporix-cli/commands/segments#epx-segments-items-delete) |
| `PUT /customer-segment/{tenant}/segments/{segmentId}/items/{type}/bulk` | Built | [`epx segments items bulk-update`](/emporix-cli/commands/segments#epx-segments-items-bulk-update) |
| `DELETE /customer-segment/{tenant}/segments/{segmentId}/items/{type}/bulk` | Built | [`epx segments items bulk-delete`](/emporix-cli/commands/segments#epx-segments-items-bulk-delete) |

All services: [Emporix API coverage](/emporix-cli/coverage).
