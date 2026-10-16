package cli

// iamReads returns the IAM Service reads.
func iamReads() []readCase {
	return []readCase{
		{
			// Users are read as one page of 500: Emporix orders them differently on every request.
			args:   []string{"iam", "users", "list"},
			path:   "/iam/acme1/users",
			scopes: []string{"iam.user_read"},
			mode:   getAllInOnePage,
			page:   "500",
			body: `[{"id":"u1","firstName":"Jane","lastName":"Doe","contactEmail":"jane@example.com","status":"ACTIVE",
				"isAccountLocked":false,"groupIds":["g1","g2"]},{"id":"u2","status":"INACTIVE","isAccountLocked":true}]`,
			want: "ID  NAME      EMAIL             STATUS    LOCKED  GROUPS\n" +
				"u1  Jane Doe  jane@example.com  ACTIVE    false   2\n" +
				"u2                              INACTIVE  true    0\n",
		},
		{
			args:   []string{"iam", "users", "get", "u/1"},
			path:   "/iam/acme1/users/u%2F1",
			scopes: []string{"iam.user_read"},
			mode:   getWhole,
			body:   `{"id":"u/1","firstName":"Jane","lastName":"Doe","status":"ACTIVE","groupIds":["g1"]}`,
			want:   "ID   NAME      EMAIL  STATUS  LOCKED  GROUPS\nu/1  Jane Doe         ACTIVE          1\n",
		},
		{
			// A user's groups need the group scope only.
			args:   []string{"iam", "users", "groups", "list", "u1", "--language", "de"},
			path:   "/iam/acme1/users/u1/groups",
			scopes: []string{"iam.group_read"},
			body: `[{"id":"g1","code":"BO_ADMIN","name":{"en":"Administrators","de":"Administratoren"},
				"userType":"EMPLOYEE","accessControls":["a1","a2"]}]`,
			want: "ID  CODE      NAME             TYPE      ACCESS CONTROLS\n" +
				"g1  BO_ADMIN  Administratoren  EMPLOYEE  2\n",
		},
		{
			// The scopes come as one text; they are sorted, without the tenant entry.
			args:   []string{"iam", "users", "scopes", "u1"},
			path:   "/iam/acme1/users/u1/scopes",
			scopes: []string{"iam.scope_read"},
			mode:   getWhole,
			body:   `{"userId":"u1","scopes":"tenant=acme1 product.product_read  order.order_read"}`,
			want:   "SCOPE\norder.order_read\nproduct.product_read\n",
		},
		{
			args:    []string{"iam", "groups", "list", "--language", "en"},
			path:    "/iam/acme1/groups",
			scopes:  []string{"iam.group_read"},
			filters: listFilters{kinds: filterable},
			body: `[{"id":"g1","code":"CUSTOMER","name":{"en":"Customers"},"userType":"CUSTOMER","accessControls":[]},
				{"id":"g2","code":"BO_USER","name":{"en":"Users"},"userType":"EMPLOYEE"}]`,
			want: "ID  CODE      NAME       TYPE      ACCESS CONTROLS\n" +
				"g1  CUSTOMER  Customers  CUSTOMER  0\n" +
				"g2  BO_USER   Users      EMPLOYEE  0\n",
		},
		{
			args:   []string{"iam", "groups", "get", "g1", "--language", "en"},
			path:   "/iam/acme1/groups/g1",
			scopes: []string{"iam.group_read"},
			mode:   getWhole,
			body:   `{"id":"g1","code":"CUSTOMER","name":{"en":"Customers"},"userType":"CUSTOMER"}`,
			want:   "ID  CODE      NAME       TYPE      ACCESS CONTROLS\ng1  CUSTOMER  Customers  CUSTOMER  0\n",
		},
		{
			// A group's members need the user scope only.
			args:   []string{"iam", "groups", "users", "list", "g1"},
			path:   "/iam/acme1/groups/g1/users",
			scopes: []string{"iam.user_read"},
			body:   `[{"id":"m1","groupId":"g1","userId":"u1","userType":"employee"}]`,
			want:   "ID  USER  TYPE\nm1  u1    employee\n",
		},
	}
}

// iamAccessReads returns the IAM Service reads of roles, permissions, scopes,
// resources, access controls and templates. Their lists sort by id unless
// --sort says otherwise.
func iamAccessReads() []readCase {
	return []readCase{
		{
			args: []string{"iam", "roles", "list", "--language", "de"}, path: "/iam/acme1/roles", query: "sort=id%3Aasc",
			scopes: []string{"iam.role_read"}, filters: listFilters{kinds: sortable},
			body: `[{"id":"r1","name":{"en":"Admin","de":"Verwalter"},"permissions":[{"id":"p1"},{"id":"p2"}]}]`,
			want: "ID  NAME       PERMISSIONS\nr1  Verwalter  2\n",
		},
		{
			args: []string{"iam", "roles", "get", "r1", "--language", "en"}, path: "/iam/acme1/roles/r1",
			scopes: []string{"iam.role_read"}, mode: getWhole, body: `{"id":"r1","name":{"en":"Admin"}}`,
			want: "ID  NAME   PERMISSIONS\nr1  Admin  0\n",
		},
		{
			args: []string{"iam", "permissions", "list", "--language", "en"}, path: "/iam/acme1/permissions",
			query: "sort=id%3Aasc", scopes: []string{"iam.permission_read"}, filters: listFilters{kinds: sortable},
			body: `[{"id":"p1","code":"PRODUCT_READ","name":{"en":"Read products"},"applicableResources":["x"]}]`,
			want: "ID  CODE          NAME           RESOURCES\np1  PRODUCT_READ  Read products  1\n",
		},
		{
			args: []string{"iam", "permissions", "get", "p1", "--language", "en"}, path: "/iam/acme1/permissions/p1",
			scopes: []string{"iam.permission_read"}, mode: getWhole,
			body: `{"id":"p1","code":"PRODUCT_READ","name":{"en":"Read products"}}`,
			want: "ID  CODE          NAME           RESOURCES\np1  PRODUCT_READ  Read products  0\n",
		},
		{
			args: []string{"iam", "scopes", "list", "--language", "de"}, path: "/iam/acme1/scopes", query: "sort=id%3Aasc",
			scopes: []string{"iam.scope_read"}, filters: listFilters{kinds: filterable},
			body: `[{"id":"product.product_read","domain":"product","predefined":true,
				"description":{"en":"Read products","de":"Produkte lesen"}},{"id":"custom.doc_read","predefined":false}]`,
			want: "ID                    DOMAIN   PREDEFINED  DESCRIPTION\n" +
				"product.product_read  product  true        Produkte lesen\n" +
				"custom.doc_read                false       \n",
		},
		{
			args: []string{"iam", "scopes", "get", "product.product_read", "--language", "en"},
			path: "/iam/acme1/scopes/product.product_read", scopes: []string{"iam.scope_read"}, mode: getWhole,
			body: `{"id":"product.product_read","domain":"product","predefined":true}`,
			want: "ID                    DOMAIN   PREDEFINED  DESCRIPTION\n" +
				"product.product_read  product  true        \n",
		},
		{
			args: []string{"iam", "resources", "list", "--language", "en"}, path: "/iam/acme1/resources",
			query: "sort=id%3Aasc", scopes: []string{"iam.resource_read"}, filters: listFilters{kinds: sortable},
			body: `[{"id":"res1","code":"PRODUCTS","name":{"en":"Products"}}]`,
			want: "ID    CODE      NAME\nres1  PRODUCTS  Products\n",
		},
		{
			args: []string{"iam", "resources", "get", "res1", "--language", "en"}, path: "/iam/acme1/resources/res1",
			scopes: []string{"iam.resource_read"}, mode: getWhole, body: `{"id":"res1","code":"PRODUCTS"}`,
			want: "ID    CODE      NAME\nres1  PRODUCTS  \n",
		},
		{
			args: []string{"iam", "access-controls", "list", "--language", "en"}, path: "/iam/acme1/access-controls",
			query: "sort=id%3Aasc", scopes: []string{"iam.access_read"}, filters: listFilters{kinds: filterable},
			body: `[{"id":"ac1","name":{"en":"Read products"},"resourceId":"res1","roleId":"r1",
				"scopes":["product.product_read"],"predefined":true}]`,
			want: "ID   NAME           SCOPES  PREDEFINED  RESOURCE  ROLE\n" +
				"ac1  Read products  1       true        res1      r1\n",
		},
		{
			args: []string{"iam", "access-controls", "get", "ac1", "--language", "en"},
			path: "/iam/acme1/access-controls/ac1", scopes: []string{"iam.access_read"}, mode: getWhole,
			body: `{"id":"ac1","resourceId":"res1","roleId":"r1"}`,
			want: "ID   NAME  SCOPES  PREDEFINED  RESOURCE  ROLE\nac1        0                   res1      r1\n",
		},
		{
			args: []string{"iam", "templates", "list", "--language", "en"}, path: "/iam/acme1/templates",
			query: "sort=id%3Aasc", scopes: []string{"iam.template_read"}, filters: listFilters{kinds: sortable},
			body: `[{"id":"t1","name":{"en":"Shop managers"},"accessControls":["ac1","ac2"]}]`,
			want: "ID  NAME           ACCESS CONTROLS\nt1  Shop managers  2\n",
		},
	}
}

// iamUserAccessReads returns the IAM Service reads of what one group or user
// may do, and of the users of a vendor.
func iamUserAccessReads() []readCase {
	return []readCase{
		{
			args: []string{"iam", "groups", "access-controls", "g1", "--language", "de"},
			path: "/iam/acme1/groups/g1/access-controls", scopes: []string{"iam.access_read"},
			body: `[{"id":"ac1","name":{"en":"Read products","de":"Produkte lesen"},"resourceId":"res1","roleId":"r1",
				"scopes":["product.product_read"],"predefined":true},{"id":"ac2","scopes":[]}]`,
			want: "ID   NAME            SCOPES  PREDEFINED  RESOURCE  ROLE\n" +
				"ac1  Produkte lesen  1       true        res1      r1\n" +
				"ac2                  0                             \n",
		},
		{
			// Emporix lists a user's access controls on one page only.
			args: []string{"iam", "users", "access-controls", "u1", "--language", "en"},
			path: "/iam/acme1/users/u1/access-controls", scopes: []string{"iam.access_read"},
			mode: getAllOnFirstPage, page: "1000",
			body: `[{"id":"ac1","name":{"en":"Read products"},
				"scopes":["product.product_read","product.product_read_unpublished"],"predefined":false}]`,
			want: "ID   NAME           SCOPES  PREDEFINED  RESOURCE  ROLE\n" +
				"ac1  Read products  2       false                 \n",
		},
		{
			// Paging parameters do not page these reads, so none are sent.
			args: []string{"iam", "users", "resource-access-controls", "u1", "res 1", "--language", "en"},
			path: "/iam/acme1/users/u1/access-controls/res%201", scopes: []string{"iam.access_read"}, mode: getWhole,
			body: `[{"id":"ac1","name":{"en":"Read products"},"resourceId":"res 1","roleId":"r1","role":{"id":"r1"},
				"resource":{"id":"res 1"},"scopes":[]}]`,
			want: "ID   NAME           SCOPES  PREDEFINED  RESOURCE  ROLE\n" +
				"ac1  Read products  0                   res 1     r1\n",
		},
		{
			args: []string{"iam", "users", "resource-permissions", "u1", "res1", "--language", "en"},
			path: "/iam/acme1/users/u1/permissions/res1", scopes: []string{"iam.permission_read"}, mode: getWhole,
			body: `[{"id":"p1","code":"PRODUCT_READ","name":{"en":"Read products"},"applicableResources":["res1","res2"]}]`,
			want: "ID  CODE          NAME           RESOURCES\np1  PRODUCT_READ  Read products  2\n",
		},
		{
			args: []string{"iam", "users", "groups", "get", "u/1", "g1", "--language", "en"},
			path: "/iam/acme1/users/u%2F1/groups/g1", scopes: []string{"iam.group_read"}, mode: getWhole,
			body: `{"id":"g1","code":"BO_ADMIN","name":{"en":"Administrators"},"userType":"EMPLOYEE","accessControls":["a1"]}`,
			want: "ID  CODE      NAME            TYPE      ACCESS CONTROLS\ng1  BO_ADMIN  Administrators  EMPLOYEE  1\n",
		},
		{
			// The reference documents no paging for a vendor's users: they are paged, checked, with GetAll.
			args: []string{"iam", "users", "for-vendor", "v1"},
			path: "/iam/acme1/users/vendors/v1", scopes: []string{"iam.user_read"},
			body: `[{"id":"u1","firstName":"Jane","lastName":"Doe","contactEmail":"jane@example.com","status":"ACTIVE",
				"isAccountLocked":false,"groupIds":["g1"]}]`,
			want: "ID  NAME      EMAIL             STATUS  LOCKED  GROUPS\nu1  Jane Doe  jane@example.com  ACTIVE  false   1\n",
		},
	}
}
