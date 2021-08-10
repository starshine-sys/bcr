module github.com/starshine-sys/bcr

go 1.15

require (
	emperror.dev/errors v0.8.0
	github.com/ReneKroon/ttlcache/v2 v2.1.0
	github.com/diamondburned/arikawa/v3 v3.0.0-20210810210230-f7880b91ee2f
	github.com/spf13/pflag v1.0.5
	github.com/starshine-sys/snowflake/v2 v2.0.0
	go.uber.org/zap v1.16.0
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
)

replace github.com/diamondburned/arikawa/v3 => ../arikawa
