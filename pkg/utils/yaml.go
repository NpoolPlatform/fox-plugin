package utils

import (
	"encoding/json"
)

type Entry map[string]interface{}

func CoverEntry(temp, in Entry) Entry {
	for k, v := range in {
		if _, ok := in[k]; ok {
			temp[k] = v
		}
	}
	return temp
}

func StructToMap(in interface{}) (Entry, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	ret := make(Entry)
	err = json.Unmarshal(out, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func MapToStruct(in Entry, out interface{}) error {
	inBytes, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(inBytes, out)
}
