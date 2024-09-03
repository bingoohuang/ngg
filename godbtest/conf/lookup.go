package conf

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/bingoohuang/ngg/ss"
	"github.com/imroc/req/v3"
)

// https://misfra.me/2023/ctes-as-lookup-tables/
/*
	   sqlite> SELECT code FROM data;
	   +------+
	   | code |
	   +------+
	   | us   |
	   | fr   |
	   | in   |
	   +------+
	One approach is to use a CASE expression in your query like this:

	sqlite> SELECT code,
	   ...> CASE code
	   ...>   WHEN 'us' THEN 'United States'
	   ...>   WHEN 'fr' THEN 'France'
	   ...>   WHEN 'in' THEN 'India'
	   ...>  END AS country
	   ...> FROM data;
	+------+---------------+
	| code |    country    |
	+------+---------------+
	| us   | United States |
	| fr   | France        |
	| in   | India         |
	+------+---------------+
	The downside is that it’s harder to read, and if you need to do something similar elsewhere in the query, you’ll have to repeat the expression.

	An alternative is to use a CTE like this:

	sqlite> WITH countries (code, name) AS (
	   ...>   SELECT * FROM (VALUES
	   ...>     ('us', 'United States'), ('fr', 'France'), ('in', 'India')
	   ...>   ) AS codes
	   ...> )
	   ...> SELECT data.code, name FROM data LEFT JOIN countries ON countries.code = data.code;
	+------+---------------+
	| code |     name      |
	+------+---------------+
	| us   | United States |
	| fr   | France        |
	| in   | India         |
	+------+---------------+
	Now the countries CTE becomes a lookup table that you can reference several times in your queries.

	This approach works with SQLite and PostgreSQL.
*/

func init() {
	registerOptions(`%lookup`, `%lookup on/off;
%lookup clear [all/name];
%lookup table.name,table.name2 '{"us": "United States", "fr": "France", "in": "India"}';
%lookup table.name ./table.name.mapping.json;
`,
		func(name string, options *replOptions) {
		}, lookupSet)
}

func lookupSet(name string, options *replOptions, args []string, pureArg string) error {
	if len(args) == 0 {
		return nil
	}

	arg0 := args[0]
	switch arg0 {
	case "on":
		options.lookupsOn = true
	case "off":
		options.lookupsOn = false
	case "clear":
		if len(args) > 1 {
			arg1 := args[1]
			switch arg1 {
			case "all":
				options.lookups = map[string]map[string]string{}
			default:
				if _, ok := options.lookups[arg1]; !ok {
					return fmt.Errorf("%s not found in lookup map", arg1)
				} else {
					delete(options.lookups, arg1)
				}
			}
		} else {
			options.lookups = map[string]map[string]string{}
		}
	default:
		if len(args) > 1 {
			arg1 := args[1]
			jsonData, err := func() ([]byte, error) {
				if ok, _ := ss.Exists(arg1); ok {
					jsonData, err := os.ReadFile(arg1)
					if err != nil {
						return nil, fmt.Errorf("read %s: %w", arg1, err)
					}
					return jsonData, nil
				}

				if httpURL, err := url.Parse(arg1); err == nil && ss.AnyOf(httpURL.Scheme, "http", "https") {
					if rsp, err := req.R().Get(arg1); err != nil {
						return nil, fmt.Errorf("rest %v: %w", arg1, err)
					} else {
						return rsp.Bytes(), nil
					}
				}

				return []byte(arg1), nil
			}()
			if err != nil {
				return err
			}

			var mapping map[string]string
			if err := json.Unmarshal(jsonData, &mapping); err != nil {
				return fmt.Errorf("unmarshal json %s: %w", arg1, err)
			}
			tableNames := strings.ToLower(arg0)
			tableNamesSlice := ss.Split(tableNames, ",")
			for _, tableName := range tableNamesSlice {
				options.lookups[tableName] = mapping
			}

			options.lookupsOn = true
		}
	}

	return nil
}
