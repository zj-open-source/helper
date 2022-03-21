package datatypes

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

func ParseAddress(text string) (*Address, error) {
	u, err := url.Parse(text)
	if err != nil {
		return nil, err
	}

	a := &Address{}

	if u.Scheme != "asset" {
		a.URL = u.String()
		return a, nil
	}

	a.Group = u.Hostname()

	if len(u.Path) > 0 {
		a.Key = u.Path[1:]
		extLastIndex := strings.LastIndex(u.Path, ".")

		if extLastIndex != -1 {
			a.Ext = u.Path[extLastIndex+1:]
			a.Key = u.Path[1:extLastIndex]
		}
	}
	return a, nil
}

// openapi:strfmt address
type Address struct {
	URL   string
	Group string
	Key   string
	Ext   string
}

func (asset Address) MarshalText() ([]byte, error) {
	return []byte(asset.String()), nil
}

func (a *Address) UnmarshalText(text []byte) (err error) {
	address, err := ParseAddress(string(text))
	if err != nil {
		return err
	}
	*a = *address
	return nil

}

func (a Address) String() string {
	if a.URL != "" {
		return a.URL
	}

	if a.Group == "" && a.Key == "" {
		return ""
	}

	u := fmt.Sprintf("asset://%s/%s", a.Group, a.Key)
	if a.Ext != "" {
		return fmt.Sprintf("%s.%s", u, a.Ext)
	}
	return u
}

func (a Address) Path() string {
	if a.URL != "" {
		return a.URL
	}

	if a.Group == "" && a.Key == "" {
		return ""
	}

	u := fmt.Sprintf("%s/%s", a.Group, a.Key)
	if a.Ext != "" {
		return fmt.Sprintf("%s.%s", u, a.Ext)
	}
	return u
}

func (a Address) DataType(string) string {
	return "varchar(1024)"
}

func (a Address) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *Address) Scan(src interface{}) error {
	return a.UnmarshalText([]byte(src.(string)))
}
