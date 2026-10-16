# Customer Service (customer-managed)

> Which Customer Service (customer-managed) operations epx sends, which do not apply, and which are planned.

Operations of the Emporix API reference file [`companies-and-customers/customer-management/api-reference/api.yml`](https://github.com/emporix/api-references/blob/faad61403b0683a7a8948b571f10305393760ae0/companies-and-customers/customer-management/api-reference/api.yml) at commit [`faad614`](https://github.com/emporix/api-references/commit/faad61403b0683a7a8948b571f10305393760ae0) (2026-10-14).

- Last change: 2026-10-01 ([`7fad4ee`](https://github.com/emporix/api-references/commit/7fad4ee))
- Version: not stated in the reference file
- Operations: 25 (0 built, 25 not applicable, 0 planned, 0 to do)

| Operation | Status | Details |
| --- | --- | --- |
| `GET /customerlogin/auth/anonymous/login` | Not applicable | Issues an anonymous storefront token; epx authenticates with a service token. |
| `GET /customerlogin/auth/anonymous/refresh` | Not applicable | Issues an anonymous storefront token; epx authenticates with a service token. |
| `POST /customer/{tenant}/signup` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `POST /customer/{tenant}/login` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `GET /customer/{tenant}/logout` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /customer/{tenant}/refreshauthtoken` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `GET /customer/{tenant}/validateauthtoken` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /customer/{tenant}/exchangeauthtoken` | Not applicable | Exchanges an external identity provider's token for a customer's token (storefront login). |
| `POST /customer/{tenant}/socialLogin` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `PATCH /customer/{tenant}/me` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /customer/{tenant}/me` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /customer/{tenant}/me/addresses` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /customer/{tenant}/me/addresses` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `PATCH /customer/{tenant}/me/addresses/{addressId}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /customer/{tenant}/me/addresses/{addressId}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `DELETE /customer/{tenant}/me/addresses/{addressId}` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /customer/{tenant}/me/addresses/{addressId}/tags` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `DELETE /customer/{tenant}/me/addresses/{addressId}/tags` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /customer/{tenant}/me/accounts/internal/email/change` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `POST /customer/{tenant}/me/accounts/internal/email/change/confirm` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `POST /customer/{tenant}/password/reset` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `POST /customer/{tenant}/password/reset/update` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `POST /customer/{tenant}/password/change` | Not applicable | Needs a signed-in customer's token, which a service token is not. |
| `GET /customer/{tenant}/signup/optin/{token}` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |
| `POST /customer/{tenant}/signup/optin/refresh_token` | Not applicable | Needs an anonymous storefront token (customer signup and login flows), which a service token is not. |

All services: [Emporix API coverage](https://commerce-cloud-integrations.github.io/emporix-cli/coverage).
