module github.com/liftoffio/segment-io

go 1.15

require (
	github.com/klauspost/compress v1.15.9
	github.com/pierrec/lz4/v4 v4.1.15
	github.com/segmentio/kafka-go v0.4.38
	github.com/stretchr/testify v1.8.0
	github.com/xdg/scram v1.0.5
	golang.org/x/net v0.0.0-20220706163947-c90051bbdb60
)

retract (
	[v0.4.36, v0.4.37]
)
