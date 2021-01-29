package bcr

import (
	"time"

	"github.com/starshine-sys/snowflake/v2"
)

var sGen = snowflake.NewGen(time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC))

func snowflakeInSlice(s snowflake.Snowflake, slice []snowflake.Snowflake) bool {
	for _, e := range slice {
		if s == e {
			return true
		}
	}
	return false
}
