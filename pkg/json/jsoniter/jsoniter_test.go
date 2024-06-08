package jsoniter

import (
	"testing"

	stdjsoniter "github.com/json-iterator/go"
)

type FactoryA struct {
	A string `json:"a"`
}

type FactoryB struct {
	B string `json:"b"`
}

func TestJSONRawMessage(t *testing.T) {
	type Data struct {
		Type    string                 `json:"type"`
		Factory stdjsoniter.RawMessage `json:"factory"`
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
		_ = Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		_ = Unmarshal(dataValue.Factory, factoryValue)
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
		_ = Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		_ = Unmarshal(dataValue.Factory, factoryValue)
		t.Log(Stringify(factoryValue))
	default:
		t.Error("unknown factory type")
	}
}

func TestJSONAny(t *testing.T) {
	type Data struct {
		Type    string          `json:"type"`
		Factory stdjsoniter.Any `json:"factory"`
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
