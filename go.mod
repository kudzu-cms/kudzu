module github.com/bobbygryzynger/ponzu

go 1.15

require (
	content v0.0.0-00010101000000-000000000000
	github.com/blevesearch/bleve v1.0.10
	github.com/boltdb/bolt v1.3.1
	github.com/cznic/b v0.0.0-20181122101859-a26611c4d92d // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gorilla/schema v1.2.0
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/nilslice/email v0.1.0
	github.com/nilslice/jwt v1.0.0
	github.com/ponzu-cms/ponzu v0.11.0
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	github.com/tidwall/gjson v1.6.1
	github.com/tidwall/sjson v1.1.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/text v0.3.3
)

replace content => ./content
