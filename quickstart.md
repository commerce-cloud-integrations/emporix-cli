# Quickstart

> Go from an Emporix API key to your first epx reads in five minutes.

This quickstart walks you through your first reads: you check your API key, list your
sites, read products, and call an endpoint directly.

Before you begin, [install epx](/emporix-cli/install) and [create an API key](/emporix-cli/configure#create-an-api-key)
with the `product.product_read_unpublished` scope.

    ```sh
    export EMPORIX_TENANT=acme1
    export EMPORIX_CLIENT_ID=...
    export EMPORIX_CLIENT_SECRET=...
    ```

    To keep the key between sessions, store it in a
    [profile](/emporix-cli/configure#store-keys-in-profiles) instead.

    ```sh
    epx auth status --scope product.product_read_unpublished
    ```

    The output shows your tenant, the token's lifetime, and the scopes granted. If the key
    lacks the scope, epx says so and exits with code 3.

    ```sh
    epx sites list
    ```

    Reads print a table by default. Add `-o json` to print the whole response, every value
    as Emporix sends it.

    ```sh
    epx products list --limit 5
    epx products get 100200 -o json
    ```

    `--limit 5` prints at most five products, and epx stops reading pages as soon as it
    can. Tables show English text by default; pass `--language` to pick another
    translation:

    ```sh
    epx products list --limit 5 --language nl
    ```

    ```sh
    epx products list --query 'name.en:~shirt published:true' --sort code:desc --limit 20
    ```

    `--query` takes an Emporix q filter and `--sort` an Emporix sort order, both passed as
    given.

    ```sh
    epx api get '/product/{tenant}/products' -p pageSize=5 --scope product.product_read_unpublished
    ```

    `epx api get` sends a GET request to any Emporix API path and prints the JSON
    response. epx replaces `{tenant}` with your tenant; quote the path so your shell
    leaves the braces alone.

## Next steps

    Tables, CSV, JSON, YAML, and JSON Lines.
    Which lists take which flags, and how paging works.
    Send your first write with a dry run.
    Every command, its flags, and its scopes.
