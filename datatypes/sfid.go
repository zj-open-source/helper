package datatypes

import (
	"strconv"
)

// deprecated
type UUID = SFID

// openapi:strfmt snowflake-id
type SFID uint64

func (sfID SFID) MarshalText() ([]byte, error) {
	return []byte(sfID.String()), nil
}

func (sfID *SFID) UnmarshalText(data []byte) (err error) {
	str := string(data)
	if len(str) == 0 {
		return
	}
	var u uint64
	u, err = strconv.ParseUint(str, 10, 64)
	*sfID = SFID(u)
	return
}

func (sfID SFID) String() string {
	return strconv.FormatUint(uint64(sfID), 10)
}

// deprecated
type UUIDs = SFIDs

type SFIDs []SFID

func (uuids SFIDs) ToUint64() []uint64 {
	var l []uint64
	for _, id := range uuids {
		l = append(l, uint64(id))
	}
	return l
}
