package simplejson

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	jsoniter "github.com/json-iterator/go"
)

func digestObject(stream *jsoniter.Stream, top jsoniter.Any) error {
	stream.WriteObjectStart()
	newObjKeys := top.Keys()
	for idx, key := range newObjKeys {
		if idx != 0 {
			stream.Write([]byte{','})
		}
		stream.WriteObjectField(key)
		val := top.Get(key)
		switch val.ValueType() {
		case jsoniter.InvalidValue:
			return errors.New("invalid value exists")
		case jsoniter.NumberValue, jsoniter.NilValue, jsoniter.BoolValue:
			val.WriteTo(stream)
		case jsoniter.StringValue:
			v := val.ToString()
			if utf8.RuneCountInString(v) > 30 {
				var sb strings.Builder
				var cnt int
				for len(v) > 0 {
					if cnt == 30 {
						sb.WriteString("...")
						break
					}
					r, size := utf8.DecodeRuneInString(v)
					_, err := sb.WriteRune(r)
					if err != nil {
						break
					}
					v = v[size:]
					cnt++
				}
				v = sb.String()
			}
			stream.WriteString(v)
		case jsoniter.ArrayValue:
			len := val.Size()
			if len == 0 {
				stream.WriteEmptyArray()
				break
			}
			ele0T := val.Get(0)
			ele0VType := ele0T.ValueType()
			if ele0VType != jsoniter.ObjectValue {
				val.WriteTo(stream)
			} else {
				stream.WriteArrayStart()
				for i := 0; i < len; i++ {
					eleT := val.Get(i)
					if err := digestObject(stream, eleT); err != nil {
						return err
					}
				}
				stream.WriteArrayEnd()
			}
		case jsoniter.ObjectValue:
			if err := digestObject(stream, val); err != nil {
				return err
			}
		}
	}
	stream.WriteObjectEnd()
	return nil
}

func DigestToJson(m any) (string, error) {
	mJson, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	mIter := jsoniter.ParseBytes(jsoniter.ConfigCompatibleWithStandardLibrary, mJson)
	if err := mIter.Error; err != nil {
		return "", err
	}
	mTop := mIter.ReadAny()
	stream := jsoniter.NewStream(jsoniter.ConfigCompatibleWithStandardLibrary, nil, len(mJson)/2)
	mVType := mTop.ValueType()
	if mVType != jsoniter.ObjectValue {
		return "", errors.New("input must be an object")
	}
	if err := digestObject(stream, mTop); err != nil {
		return "", err
	}
	return string(stream.Buffer()), nil
}
