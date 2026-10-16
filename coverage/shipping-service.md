# Shipping Service

> Which Shipping Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`delivery-and-shipping/shipping/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/delivery-and-shipping/shipping/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-09-11 ([`eeddb11`](https://github.com/emporix/api-references/commit/eeddb11))
- Version: not stated in the reference file
- Operations: 45 (44 built, 1 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /shipping/{tenant}/areaDeliveryTimes/{deliveryAreaId}/{cartId}` | Not applicable | Deprecated by Emporix in 2023; `epx delivery-times windows` reads a cart's delivery windows. |
| `GET /shipping/{tenant}/actualDeliveryWindows/{cartId}` | Built | [`epx delivery-times windows`](/emporix-cli/commands/delivery-times#epx-delivery-times-windows) |
| `POST /shipping/{tenant}/actualDeliveryWindows/incrementCounter` | Built | [`epx delivery-times increment-counter`](/emporix-cli/commands/delivery-times#epx-delivery-times-increment-counter) |
| `POST /shipping/{tenant}/findSite` | Built | [`epx shipping find-sites`](/emporix-cli/commands/shipping#epx-shipping-find-sites) |
| `POST /shipping/{tenant}/{site}/zones` | Built | [`epx shipping zones create`](/emporix-cli/commands/shipping#epx-shipping-zones-create) |
| `GET /shipping/{tenant}/{site}/zones` | Built | [`epx shipping zones list`](/emporix-cli/commands/shipping#epx-shipping-zones-list) |
| `GET /shipping/{tenant}/{site}/zones/{zoneId}` | Built | [`epx shipping zones get`](/emporix-cli/commands/shipping#epx-shipping-zones-get) |
| `PUT /shipping/{tenant}/{site}/zones/{zoneId}` | Built | [`epx shipping zones update`](/emporix-cli/commands/shipping#epx-shipping-zones-update) |
| `PATCH /shipping/{tenant}/{site}/zones/{zoneId}` | Built | [`epx shipping zones patch`](/emporix-cli/commands/shipping#epx-shipping-zones-patch) |
| `DELETE /shipping/{tenant}/{site}/zones/{zoneId}` | Built | [`epx shipping zones delete`](/emporix-cli/commands/shipping#epx-shipping-zones-delete) |
| `POST /shipping/{tenant}/deliveryWindowValidation` | Built | [`epx delivery-times validate`](/emporix-cli/commands/delivery-times#epx-delivery-times-validate) |
| `POST /shipping/{tenant}/{site}/zones/{zoneId}/methods` | Built | [`epx shipping zones methods create`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-create) |
| `GET /shipping/{tenant}/{site}/zones/{zoneId}/methods` | Built | [`epx shipping zones methods list`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-list) |
| `GET /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | Built | [`epx shipping zones methods get`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-get) |
| `PUT /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | Built | [`epx shipping zones methods update`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-update) |
| `PATCH /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | Built | [`epx shipping zones methods patch`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-patch) |
| `DELETE /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` | Built | [`epx shipping zones methods delete`](/emporix-cli/commands/shipping#epx-shipping-zones-methods-delete) |
| `POST /shipping/{tenant}/{site}/quote` | Built | [`epx shipping quote`](/emporix-cli/commands/shipping#epx-shipping-quote) |
| `POST /shipping/{tenant}/{site}/quote/minimum` | Built | [`epx shipping quote-minimum`](/emporix-cli/commands/shipping#epx-shipping-quote-minimum) |
| `POST /shipping/{tenant}/{site}/quote/slot` | Built | [`epx shipping quote-slot`](/emporix-cli/commands/shipping#epx-shipping-quote-slot) |
| `POST /shipping/{tenant}/{site}/groups` | Built | [`epx shipping groups create`](/emporix-cli/commands/shipping#epx-shipping-groups-create) |
| `GET /shipping/{tenant}/{site}/groups` | Built | [`epx shipping groups list`](/emporix-cli/commands/shipping#epx-shipping-groups-list) |
| `GET /shipping/{tenant}/{site}/groups/{groupId}` | Built | [`epx shipping groups get`](/emporix-cli/commands/shipping#epx-shipping-groups-get) |
| `PUT /shipping/{tenant}/{site}/groups/{groupId}` | Built | [`epx shipping groups update`](/emporix-cli/commands/shipping#epx-shipping-groups-update) |
| `DELETE /shipping/{tenant}/{site}/groups/{groupId}` | Built | [`epx shipping groups delete`](/emporix-cli/commands/shipping#epx-shipping-groups-delete) |
| `POST /shipping/{tenant}/{site}/cgrelations` | Built | [`epx shipping customers create`](/emporix-cli/commands/shipping#epx-shipping-customers-create) |
| `GET /shipping/{tenant}/{site}/cgrelations` | Built | [`epx shipping customers list`](/emporix-cli/commands/shipping#epx-shipping-customers-list) |
| `GET /shipping/{tenant}/{site}/cgrelations/{customerId}` | Built | [`epx shipping customers get`](/emporix-cli/commands/shipping#epx-shipping-customers-get) |
| `PUT /shipping/{tenant}/{site}/cgrelations/{customerId}` | Built | [`epx shipping customers update`](/emporix-cli/commands/shipping#epx-shipping-customers-update) |
| `DELETE /shipping/{tenant}/{site}/cgrelations/{customerId}` | Built | [`epx shipping customers delete`](/emporix-cli/commands/shipping#epx-shipping-customers-delete) |
| `GET /shipping/{tenant}/delivery-times` | Built | [`epx delivery-times list`](/emporix-cli/commands/delivery-times#epx-delivery-times-list) |
| `POST /shipping/{tenant}/delivery-times` | Built | [`epx delivery-times create`](/emporix-cli/commands/delivery-times#epx-delivery-times-create) |
| `POST /shipping/{tenant}/delivery-times/bulk` | Built | [`epx delivery-times bulk-create`](/emporix-cli/commands/delivery-times#epx-delivery-times-bulk-create) |
| `GET /shipping/{tenant}/delivery-times/{deliveryTimeId}` | Built | [`epx delivery-times get`](/emporix-cli/commands/delivery-times#epx-delivery-times-get) |
| `PUT /shipping/{tenant}/delivery-times/{deliveryTimeId}` | Built | [`epx delivery-times update`](/emporix-cli/commands/delivery-times#epx-delivery-times-update) |
| `PATCH /shipping/{tenant}/delivery-times/{deliveryTimeId}` | Built | [`epx delivery-times patch`](/emporix-cli/commands/delivery-times#epx-delivery-times-patch) |
| `DELETE /shipping/{tenant}/delivery-times/{deliveryTimeId}` | Built | [`epx delivery-times delete`](/emporix-cli/commands/delivery-times#epx-delivery-times-delete) |
| `GET /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots` | Built | [`epx delivery-times slots list`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-list) |
| `POST /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots` | Built | [`epx delivery-times slots create`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-create) |
| `DELETE /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots` | Built | [`epx delivery-times slots delete-all`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-delete-all) |
| `GET /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | Built | [`epx delivery-times slots get`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-get) |
| `PUT /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | Built | [`epx delivery-times slots update`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-update) |
| `PATCH /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | Built | [`epx delivery-times slots patch`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-patch) |
| `DELETE /shipping/{tenant}/delivery-times/{deliveryTimeId}/slots/{slotId}` | Built | [`epx delivery-times slots delete`](/emporix-cli/commands/delivery-times#epx-delivery-times-slots-delete) |
| `POST /shipping/{tenant}/delivery-cycles/generate` | Built | [`epx delivery-times generate-cycle`](/emporix-cli/commands/delivery-times#epx-delivery-times-generate-cycle) |

All services: [Emporix API coverage](/emporix-cli/coverage).
