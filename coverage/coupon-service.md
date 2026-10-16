# Coupon Service

> Which Coupon Service operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`rewards-and-promotions/coupon/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/rewards-and-promotions/coupon/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-09-25 ([`86c5409`](https://github.com/emporix/api-references/commit/86c5409))
- Version: not stated in the reference file
- Operations: 13 (13 built, 0 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /coupon/{tenant}/coupons` | Built | [`epx coupons list`](/emporix-cli/commands/coupons#epx-coupons-list) |
| `POST /coupon/{tenant}/coupons` | Built | [`epx coupons create`](/emporix-cli/commands/coupons#epx-coupons-create) |
| `GET /coupon/{tenant}/coupons/{code}` | Built | [`epx coupons get`](/emporix-cli/commands/coupons#epx-coupons-get) |
| `PUT /coupon/{tenant}/coupons/{code}` | Built | [`epx coupons update`](/emporix-cli/commands/coupons#epx-coupons-update) |
| `PATCH /coupon/{tenant}/coupons/{code}` | Built | [`epx coupons patch`](/emporix-cli/commands/coupons#epx-coupons-patch) |
| `DELETE /coupon/{tenant}/coupons/{code}` | Built | [`epx coupons delete`](/emporix-cli/commands/coupons#epx-coupons-delete) |
| `POST /coupon/{tenant}/coupons/{code}/validation` | Built | [`epx coupons validate`](/emporix-cli/commands/coupons#epx-coupons-validate) |
| `GET /coupon/{tenant}/coupons/{code}/redemptions` | Built | [`epx coupons redemptions`](/emporix-cli/commands/coupons#epx-coupons-redemptions) |
| `POST /coupon/{tenant}/coupons/{code}/redemptions` | Built | [`epx coupons redeem`](/emporix-cli/commands/coupons#epx-coupons-redeem) |
| `GET /coupon/{tenant}/coupons/{code}/redemptions/{id}` | Built | [`epx coupons redemption`](/emporix-cli/commands/coupons#epx-coupons-redemption) |
| `DELETE /coupon/{tenant}/coupons/{code}/redemptions/{id}` | Built | [`epx coupons redemption-delete`](/emporix-cli/commands/coupons#epx-coupons-redemption-delete) |
| `GET /coupon/{tenant}/referral-coupons/{customerNumber}` | Built | [`epx coupons referral`](/emporix-cli/commands/coupons#epx-coupons-referral) |
| `POST /coupon/{tenant}/referral-coupons/{customerNumber}` | Built | [`epx coupons referral-create`](/emporix-cli/commands/coupons#epx-coupons-referral-create) |

All services: [Emporix API coverage](/emporix-cli/coverage).
