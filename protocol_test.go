package skafka

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestApiVersionsFormat(t *testing.T) {
	for _, test := range []struct {
		version ApiVersion
		format  string
		output  string
	}{
		{version: ApiVersion{1, 2, 5}, format: "%s", output: "Fetch"},
		{version: ApiVersion{1, 2, 5}, format: "%d", output: "1"},
		{version: ApiVersion{1, 2, 5}, format: "%-d", output: "2"},
		{version: ApiVersion{1, 2, 5}, format: "%+d", output: "5"},
		{version: ApiVersion{1, 2, 5}, format: "%v", output: "Fetch[v2:v5]"},
		{version: ApiVersion{1, 2, 5}, format: "%-v", output: "v2"},
		{version: ApiVersion{1, 2, 5}, format: "%+v", output: "v5"},
		{version: ApiVersion{1, 2, 5}, format: "%#v", output: "kafka.ApiVersion{ApiKey:1 MinVersion:2 MaxVersion:5}"},
	} {
		t.Run(test.output, func(t *testing.T) {
			if s := fmt.Sprintf(test.format, test.version); s != test.output {
				t.Error("output mismatch:", s, "!=", test.output)
			}
		})
	}
}

func TestProtocol(t *testing.T) {
	tests := []interface{}{
		int8(42),
		int16(42),
		int32(42),
		int64(42),
		"",
		"Hello World!",
		[]byte(nil),
		[]byte("Hello World!"),

		requestHeader{
			Size:          26,
			ApiKey:        int16(offsetCommit),
			ApiVersion:    int16(v2),
			CorrelationID: 42,
			ClientID:      "Hello World!",
		},

		message{
			MagicByte: 1,
			Timestamp: 42,
			Key:       nil,
			Value:     []byte("Hello World!"),
		},

		topicMetadataRequestV1{"A", "B", "C"},

		metadataResponseV1{
			Brokers: []brokerMetadataV1{
				{NodeID: 1, Host: "localhost", Port: 9001},
				{NodeID: 2, Host: "localhost", Port: 9002, Rack: "rack2"},
			},
			ControllerID: 2,
			Topics: []topicMetadataV1{
				{TopicErrorCode: 0, Internal: true, Partitions: []partitionMetadataV1{{
					PartitionErrorCode: 0,
					PartitionID:        1,
					Leader:             2,
					Replicas:           []int32{1},
					Isr:                []int32{1},
				}}},
			},
		},

		topicMetadataRequestV6{
			Topics:                 []string{"A", "B", "C"},
			AllowAutoTopicCreation: true,
		},

		metadataResponseV6{
			Brokers: []brokerMetadataV1{
				{NodeID: 1, Host: "localhost", Port: 9001},
				{NodeID: 2, Host: "localhost", Port: 9002, Rack: "rack2"},
			},
			ClusterId:    "cluster",
			ControllerID: 2,
			Topics: []topicMetadataV6{
				{TopicErrorCode: 0, Internal: true, Partitions: []partitionMetadataV6{{
					PartitionErrorCode: 0,
					PartitionID:        1,
					Leader:             2,
					Replicas:           []int32{1},
					Isr:                []int32{1},
					OfflineReplicas:    []int32{1},
				}}},
			},
		},

		listOffsetRequestV1{
			ReplicaID: 1,
			Topics: []listOffsetRequestTopicV1{
				{TopicName: "A", Partitions: []listOffsetRequestPartitionV1{
					{Partition: 0, Time: -1},
					{Partition: 1, Time: -1},
					{Partition: 2, Time: -1},
				}},
				{TopicName: "B", Partitions: []listOffsetRequestPartitionV1{
					{Partition: 0, Time: -2},
				}},
				{TopicName: "C", Partitions: []listOffsetRequestPartitionV1{
					{Partition: 0, Time: 42},
				}},
			},
		},

		listOffsetResponseV1{
			{TopicName: "A", PartitionOffsets: []partitionOffsetV1{
				{Partition: 0, Timestamp: 42, Offset: 1},
			}},
			{TopicName: "B", PartitionOffsets: []partitionOffsetV1{
				{Partition: 0, Timestamp: 43, Offset: 10},
				{Partition: 1, Timestamp: 44, Offset: 100},
			}},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test), func(t *testing.T) {
			b := &bytes.Buffer{}
			r := bufio.NewReader(b)
			w := &writeBuffer{w: b}
			w.write(test)

			if size := int(sizeof(test)); size != b.Len() {
				t.Error("invalid size:", size, "!=", b.Len())
			}

			v := reflect.New(reflect.TypeOf(test))
			n := b.Len()

			n, err := read(r, n, v.Interface())
			if err != nil {
				t.Fatal(err)
			}
			if n != 0 {
				t.Errorf("%d unread bytes", n)
			}

			if !reflect.DeepEqual(test, v.Elem().Interface()) {
				t.Error("values don't match:")
				t.Logf("expected: %#v", test)
				t.Logf("found:    %#v", v.Elem().Interface())
			}
		})
	}
}
