package util_test

import (
	"encoding/json"
	"log"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
)

// go test -v -run ^TestDigestToJson$ -count=1 ./internal/app/util/json_test.go
func TestDigestToJson(t *testing.T) {
	m := struct {
		Field1 string `json:"field1"`
		Field2 bool   `json:"field2"`
		Field3 []int  `json:"field3"`
		Field4 []struct {
			Field4_field1 int  `json:"field4_field1"`
			Field4_field2 bool `json:"Field3_field2"`
		} `json:"field4"`
		Field5 int     `json:"field5"`
		Field6 float32 `json:"field6"`
		Field7 struct {
			Field7_field1 string `json:"field7_field1"`
		}
	}{
		Field1: "012345678901234567890123456789012345",
		Field2: true,
		Field3: []int{0, 1, 2, 3, 4},
		Field4: []struct {
			Field4_field1 int  `json:"field4_field1"`
			Field4_field2 bool `json:"Field3_field2"`
		}{{Field4_field1: 123, Field4_field2: true}, {Field4_field1: 456, Field4_field2: false}},
		Field5: 321,
		Field6: 654.123,
		Field7: struct {
			Field7_field1 string `json:"field7_field1"`
		}{Field7_field1: "small string"},
	}
	res, err := util.DigestToJson(m)
	if err != nil {
		t.Error(err)
	}
	log.Println(res)
}

func TestJsonNullValue(t *testing.T) {
	m := struct {
		Field1 string `json:"field1"`
		Field2 bool   `json:"field2"`
		Field3 []int  `json:"field3"`
		Field4 interface{}
	}{}
	res, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%s\n", res)
}

func TestJsonGet(t *testing.T) {
	data := []byte(`{"key1": "val1", "key2": {"key3":"val3"}, "key4":null, "key5":{}}`)
	log.Printf("%v\n", jsoniter.Get(data, "key4").GetInterface())
}
