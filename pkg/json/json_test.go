package json

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
)

type FactoryA struct {
	A string `json:"a"`
}

type FactoryB struct {
	B string `json:"b"`
}

func TestJSONRawMessage(t *testing.T) {
	type Data struct {
		Type    string              `json:"type"`
		Factory jsoniter.RawMessage `json:"factory"`
	}
	data := `{
		"type": "a",
		"factory": {
			"a": "FactoryA"
		}
	}`
	dataValue := &Data{}
	if err := UnmarshalFromString(data, dataValue); err != nil {
		t.Error(err)
		return
	}
	t.Log(Stringify(dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	default:
		t.Error("unknown factory type")
	}

	data = `{
		"type": "b",
		"factory": {
			"b": "FactoryB"
		}
	}`
	if err := UnmarshalFromString(data, dataValue); err != nil {
		t.Error(err)
		return
	}
	t.Log(Stringify(dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	default:
		t.Error("unknown factory type")
	}
}

func TestJSONAny(t *testing.T) {
	type Data struct {
		Type    string       `json:"type"`
		Factory jsoniter.Any `json:"factory"`
	}
	data := `{
		"type": "a",
		"factory": {
			"a": "FactoryA"
		}
	}`
	dataValue := &Data{}
	if err := UnmarshalFromString(data, dataValue); err != nil {
		t.Error(err)
		return
	}
	t.Log(Stringify(dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		dataValue.Factory.ToVal(factoryValue)
		t.Log(Stringify(factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		dataValue.Factory.ToVal(factoryValue)
		t.Log(Stringify(factoryValue))
	default:
		t.Error("unknown factory type")
	}

	data = `{
		"type": "b",
		"factory": {
			"b": "FactoryB"
		}
	}`
	if err := UnmarshalFromString(data, dataValue); err != nil {
		t.Error(err)
		return
	}
	t.Log(Stringify(dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		dataValue.Factory.ToVal(factoryValue)
		t.Log(Stringify(factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		dataValue.Factory.ToVal(factoryValue)
		t.Log(Stringify(factoryValue))
	default:
		t.Error("unknown factory type")
	}
}

func TestFlatten(t *testing.T) {
	data := map[string]any{
		"a": "123456",
		"b": "123456",
		"c": map[string]any{
			"a": "123456",
			"b": "123456",
		},
	}
	t.Log(Stringify(Flatten(data)))
}
