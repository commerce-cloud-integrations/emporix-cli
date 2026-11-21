package cli

func iamReads() []readCase {
	return []readCase{
		{
			args:     []string{"iam", "users", "list"},
			path:     "/iam/acme1/users",
			scopes:   []string{"iam.user_read"},
			mode:     getAllInOnePage,
			pageSize: "500",
			response: `[{"id":"u1","firstName":"Jane","lastName":"Doe","contactEmail":"jane@example.com","status":"ACTIVE",
				"isAccountLocked":false,"groupIds":["g1","g2"]},{"id":"u2","status":"INACTIVE","isAccountLocked":true}]`,
			want: "ID  NAME      EMAIL             STATUS    LOCKED  GROUPS\n" +
				"u1  Jane Doe  jane@example.com  ACTIVE    false   2\n" +
				"u2                              INACTIVE  true    0\n",
		},
		{
			args:     []string{"iam", "users", "get", "u/1"},
			path:     "/iam/acme1/users/u%2F1",
			scopes:   []string{"iam.user_read"},
			mode:     getWhole,
			response: `{"id":"u/1","firstName":"Jane","lastName":"Doe","status":"ACTIVE","groupIds":["g1"]}`,
			want:     "ID   NAME      EMAIL  STATUS  LOCKED  GROUPS\nu/1  Jane Doe         ACTIVE          1\n",
		},
		{
			args:   []string{"iam", "users", "groups", "list", "u1", "--language", "de"},
			path:   "/iam/acme1/users/u1/groups",
			scopes: []string{"iam.group_read"},
			response: `[{"id":"g1","code":"BO_ADMIN","name":{"en":"Administrators","de":"Administratoren"},
				"userType":"EMPLOYEE","accessControls":["a1","a2"]}]`,
			want: "ID  CODE      NAME             TYPE      ACCESS CONTROLS\n" +
				"g1  BO_ADMIN  Administratoren  EMPLOYEE  2\n",
		},
		{
			args:     []string{"iam", "users", "scopes", "u1"},
			path:     "/iam/acme1/users/u1/scopes",
			scopes:   []string{"iam.scope_read"},
			mode:     getWhole,
			response: `{"userId":"u1","scopes":"tenant=acme1 product.product_read  order.order_read"}`,
			want:     "SCOPE\norder.order_read\nproduct.product_read\n",
		},
		{
			args:    []string{"iam", "groups", "list", "--language", "en"},
			path:    "/iam/acme1/groups",
			query:   "sort=id%3Aasc",
			scopes:  []string{"iam.group_read"},
			filters: listFilters{kinds: filterable},
			response: `[{"id":"g1","code":"CUSTOMER","name":{"en":"Customers"},"userType":"CUSTOMER","accessControls":[]},
				{"id":"g2","code":"BO_USER","name":{"en":"Users"},"userType":"EMPLOYEE"}]`,
			want: "ID  CODE      NAME       TYPE      ACCESS CONTROLS\n" +
				"g1  CUSTOMER  Customers  CUSTOMER  0\n" +
				"g2  BO_USER   Users      EMPLOYEE  0\n",
		},
		{
			args:     []string{"iam", "groups", "get", "g1", "--language", "en"},
			path:     "/iam/acme1/groups/g1",
			scopes:   []string{"iam.group_read"},
			mode:     getWhole,
			response: `{"id":"g1","code":"CUSTOMER","name":{"en":"Customers"},"userType":"CUSTOMER"}`,
			want:     "ID  CODE      NAME       TYPE      ACCESS CONTROLS\ng1  CUSTOMER  Customers  CUSTOMER  0\n",
		},
		{
			args:     []string{"iam", "groups", "users", "list", "g1"},
			path:     "/iam/acme1/groups/g1/users",
			scopes:   []string{"iam.user_read"},
			response: `[{"id":"m1","groupId":"g1","userId":"u1","userType":"employee"}]`,
			want:     "ID  USER  TYPE\nm1  u1    employee\n",
		},
	}
}

func iamAccessReads() []readCase {
	return []readCase{
		{
			args: []string{"iam", "scopes", "list", "--language", "de"}, path: "/iam/acme1/scopes", query: "sort=id%3Aasc",
			scopes: []string{"iam.scope_read"}, filters: listFilters{kinds: filterable},
			response: `[{"id":"product.product_read","domain":"product","predefined":true,
				"description":{"en":"Read products","de":"Produkte lesen"}},{"id":"custom.doc_read","predefined":false}]`,
			want: "ID                    DOMAIN   PREDEFINED  DESCRIPTION\n" +
				"product.product_read  product  true        Produkte lesen\n" +
				"custom.doc_read                false       \n",
		},
		{
			args: []string{"iam", "scopes", "get", "product.product_read", "--language", "en"},
			path: "/iam/acme1/scopes/product.product_read", scopes: []string{"iam.scope_read"}, mode: getWhole,
			response: `{"id":"product.product_read","domain":"product","predefined":true}`,
			want: "ID                    DOMAIN   PREDEFINED  DESCRIPTION\n" +
				"product.product_read  product  true        \n",
		},
		{
			args: []string{"iam", "access-controls", "list", "--language", "en"}, path: "/iam/acme1/access-controls",
			query: "sort=id%3Aasc", scopes: []string{"iam.access_read"}, filters: listFilters{kinds: filterable},
			response: `[{"id":"ac1","name":{"en":"Read products"},"scopes":["product.product_read"],"predefined":true}]`,
			want: "ID   NAME           SCOPES  PREDEFINED\n" +
				"ac1  Read products  1       true\n",
		},
		{
			args: []string{"iam", "access-controls", "get", "ac1", "--language", "en"},
			path: "/iam/acme1/access-controls/ac1", scopes: []string{"iam.access_read"}, mode: getWhole,
			response: `{"id":"ac1"}`,
			want:     "ID   NAME  SCOPES  PREDEFINED\nac1        0       \n",
		},
		{
			args: []string{"iam", "templates", "list", "--language", "en"}, path: "/iam/acme1/templates",
			query: "sort=id%3Aasc", scopes: []string{"iam.template_read"}, filters: listFilters{kinds: sortable},
			response: `[{"id":"t1","name":{"en":"Shop managers"},"accessControls":["ac1","ac2"]}]`,
			want:     "ID  NAME           ACCESS CONTROLS\nt1  Shop managers  2\n",
		},
	}
}

func iamUserAccessReads() []readCase {
	return []readCase{
		{
			args: []string{"iam", "groups", "access-controls", "g1", "--language", "de"},
			path: "/iam/acme1/groups/g1/access-controls", scopes: []string{"iam.access_read"},
			response: `[{"id":"ac1","name":{"en":"Read products","de":"Produkte lesen"},
				"scopes":["product.product_read"],"predefined":true},{"id":"ac2","scopes":[]}]`,
			want: "ID   NAME            SCOPES  PREDEFINED\n" +
				"ac1  Produkte lesen  1       true\n" +
				"ac2                  0       \n",
		},
		{
			args: []string{"iam", "users", "access-controls", "u1", "--language", "en"},
			path: "/iam/acme1/users/u1/access-controls", scopes: []string{"iam.access_read"},
			mode: getAllOnFirstPage, pageSize: "1000",
			response: `[{"id":"ac1","name":{"en":"Read products"},
				"scopes":["product.product_read","product.product_read_unpublished"],"predefined":false}]`,
			want: "ID   NAME           SCOPES  PREDEFINED\n" +
				"ac1  Read products  2       false\n",
		},
		{
			args: []string{"iam", "users", "groups", "get", "u/1", "g1", "--language", "en"},
			path: "/iam/acme1/users/u%2F1/groups/g1", scopes: []string{"iam.group_read"}, mode: getWhole,
			response: `{"id":"g1","code":"BO_ADMIN","name":{"en":"Administrators"},"userType":"EMPLOYEE",` +
				`"accessControls":["a1"]}`,
			want: "ID  CODE      NAME            TYPE      ACCESS CONTROLS\ng1  BO_ADMIN  Administrators  EMPLOYEE  1\n",
		},
		{
			args: []string{"iam", "users", "for-vendor", "v1"},
			path: "/iam/acme1/users/vendors/v1", scopes: []string{"iam.user_read"},
			response: `[{"id":"u1","firstName":"Jane","lastName":"Doe","contactEmail":"jane@example.com","status":"ACTIVE",
				"isAccountLocked":false,"groupIds":["g1"]}]`,
			want: "ID  NAME      EMAIL             STATUS  LOCKED  GROUPS\nu1  Jane Doe  jane@example.com  ACTIVE  false   1\n",
		},
	}
}
