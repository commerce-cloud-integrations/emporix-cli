# CLI reference

> Every epx command and flag, with the OAuth scopes each one requests.

This reference documents every command and flag of epx. The same text is in `epx <command> --help`.

Command-line client for the Emporix Commerce APIs.

Credentials are read from the environment:

```text
EMPORIX_TENANT, EMPORIX_CLIENT_ID, EMPORIX_CLIENT_SECRET (and optionally EMPORIX_API_URL).
```

Where none of these is set, a profile ([`epx profile add`](/emporix-cli/commands/profile#epx-profile-add)) gives them, its client secret from the
OS keychain: the profile that `--profile` names, else EPX\_PROFILE, else the default profile ([`epx profile use`](/emporix-cli/commands/profile#epx-profile-use)). The two never mix, and naming a profile while one is set is a usage error.
`--tenant` overrides the tenant of either. A profile's default site and language apply where a
command needs a site or shows translations and the flag is not given, on the profile's tenant.

Exit codes: 0 success, 1 error, 2 invalid usage or configuration,
3 authentication or missing scopes, 4 not found, 5 conflict: a write that Emporix answered with
409 Conflict, as when the item changed since it was read or exists already, or a `units update`
or `catalogs patch` with metadata.version answered with 404 Not Found, as the Unit and Catalog
services answer a stale version. A 403 Forbidden that the help of a command calls a refusal
of its input exits 1.

## Global flags

Every command takes these flags.

- `-h, --help`\
  Help for epx.
- `--profile <name>`\
  Profile to use (default from EPX\_PROFILE, else the default profile).
- `--tenant <string>`\
  Emporix tenant (default from EMPORIX\_TENANT, else the profile).
- `--timeout <duration>`\
  Timeout for each HTTP request. Default: `30s`.

## Commands

- [`epx ai`](/emporix-cli/commands/ai)\
  Read and change AI agents and their setup, chat with them, generate texts.
- [`epx api`](/emporix-cli/commands/api)\
  Send requests to any Emporix API path.
- [`epx auth`](/emporix-cli/commands/auth)\
  Check Emporix credentials.
- [`epx availability`](/emporix-cli/commands/availability)\
  Read and change availability.
- [`epx brands`](/emporix-cli/commands/brands)\
  Read, create, change and delete brands.
- [`epx carts`](/emporix-cli/commands/carts)\
  Read and change carts and their items and discounts.
- [`epx catalogs`](/emporix-cli/commands/catalogs)\
  Read and change catalogs.
- [`epx categories`](/emporix-cli/commands/categories)\
  Read and change categories and the products assigned to them.
- [`epx category-trees`](/emporix-cli/commands/category-trees)\
  Read and rebuild category trees.
- [`epx changelog`](/emporix-cli/commands/changelog)\
  Read the change history (audit log).
- [`epx completion`](/emporix-cli/commands/completion)\
  Generate the autocompletion script for the specified shell.
- [`epx configurations`](/emporix-cli/commands/configurations)\
  Read and change tenant configurations.
- [`epx contact-assignments`](/emporix-cli/commands/contact-assignments)\
  Read and change which customers are contacts of which legal entities.
- [`epx countries`](/emporix-cli/commands/countries)\
  Read countries, and activate or deactivate them.
- [`epx coupons`](/emporix-cli/commands/coupons)\
  Read, create, change and redeem coupons.
- [`epx currencies`](/emporix-cli/commands/currencies)\
  Read, create, change and delete currencies.
- [`epx custom-entities`](/emporix-cli/commands/custom-entities)\
  Read and change custom entity types and their instances.
- [`epx customer-locations`](/emporix-cli/commands/customer-locations)\
  Read and change the locations of B2B customers.
- [`epx customers`](/emporix-cli/commands/customers)\
  Read and change customers and their addresses.
- [`epx delivery-times`](/emporix-cli/commands/delivery-times)\
  Read and change delivery times and their time slots; read the delivery windows of a cart.
- [`epx exchange-rates`](/emporix-cli/commands/exchange-rates)\
  Read, create, change and delete currency exchange rates.
- [`epx fees`](/emporix-cli/commands/fees)\
  Read and change fees and which items have them.
- [`epx iam`](/emporix-cli/commands/iam)\
  Read and change IAM users, groups and access rights.
- [`epx imports`](/emporix-cli/commands/imports)\
  Read import configurations, runs and imported records; start and schedule runs.
- [`epx indexing`](/emporix-cli/commands/indexing)\
  Read and configure search index providers; reindex.
- [`epx invoice-jobs`](/emporix-cli/commands/invoice-jobs)\
  Create and read invoice jobs.
- [`epx labels`](/emporix-cli/commands/labels)\
  Read, create, change and delete labels.
- [`epx legal-entities`](/emporix-cli/commands/legal-entities)\
  Read and change legal entities (B2B companies); read their parents.
- [`epx media`](/emporix-cli/commands/media)\
  Read, upload, link and change media assets.
- [`epx orders`](/emporix-cli/commands/orders)\
  Read, create and change orders and their status.
- [`epx payment-modes`](/emporix-cli/commands/payment-modes)\
  Read, create, update and delete payment modes.
- [`epx payment-transactions`](/emporix-cli/commands/payment-transactions)\
  Read payment transactions; authorize, capture, refund and cancel payments.
- [`epx price-lists`](/emporix-cli/commands/price-lists)\
  Read and change price lists and their prices.
- [`epx price-models`](/emporix-cli/commands/price-models)\
  Read and change price models.
- [`epx prices`](/emporix-cli/commands/prices)\
  Read and change prices.
- [`epx product-templates`](/emporix-cli/commands/product-templates)\
  Read, create, update and delete product templates.
- [`epx products`](/emporix-cli/commands/products)\
  Read, create, update and delete products.
- [`epx profile`](/emporix-cli/commands/profile)\
  Manage profiles: named settings and credentials of a tenant.
- [`epx quotes`](/emporix-cli/commands/quotes)\
  Read, create and change quotes, their PDF, history and reasons.
- [`epx regions`](/emporix-cli/commands/regions)\
  Read regions.
- [`epx returns`](/emporix-cli/commands/returns)\
  Read, create, change and delete returns.
- [`epx reward-points`](/emporix-cli/commands/reward-points)\
  Read and change reward points and the ways to redeem them.
- [`epx schemas`](/emporix-cli/commands/schemas)\
  Read, create, change and delete custom schemas and schema references.
- [`epx search`](/emporix-cli/commands/search)\
  Search documents; read and change search indexes and saved searches.
- [`epx segments`](/emporix-cli/commands/segments)\
  Read and change customer segments and what is assigned to them.
- [`epx sequential-ids`](/emporix-cli/commands/sequential-ids)\
  Read, create, activate and delete sequential ID schemas; issue IDs.
- [`epx session-contexts`](/emporix-cli/commands/session-contexts)\
  Read and change session contexts.
- [`epx shipping`](/emporix-cli/commands/shipping)\
  Read and change shipping zones, methods and groups, and calculate shipping fees.
- [`epx shopping-lists`](/emporix-cli/commands/shopping-lists)\
  Read, create, replace and delete customers' shopping lists.
- [`epx sites`](/emporix-cli/commands/sites)\
  Read, create and change sites and their mixins.
- [`epx taxes`](/emporix-cli/commands/taxes)\
  Read and change tax configurations, and calculate net and gross prices.
- [`epx units`](/emporix-cli/commands/units)\
  Read and change units of measurement; read unit types and convert between units.
- [`epx vendor-locations`](/emporix-cli/commands/vendor-locations)\
  Read and change the locations of vendors.
- [`epx vendors`](/emporix-cli/commands/vendors)\
  Read and change vendors.
- [`epx version`](/emporix-cli/commands/version)\
  Print the epx version.
- [`epx webhooks`](/emporix-cli/commands/webhooks)\
  Read and change webhook configurations and subscriptions; read statistics.
