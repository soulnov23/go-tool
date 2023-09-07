package json

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
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
		log.ErrorFields("", zap.Error(err))
		return
	}
	log.DebugFields("", zap.Reflect("data", dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		Unmarshal(dataValue.Factory, factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		Unmarshal(dataValue.Factory, factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	default:
		log.Error("unknown factory type")
	}

	data = `{
		"type": "b",
		"factory": {
			"b": "FactoryB"
		}
	}`
	if err := UnmarshalFromString(data, dataValue); err != nil {
		log.ErrorFields("", zap.Error(err))
		return
	}
	log.DebugFields("", zap.Reflect("data", dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		Unmarshal(dataValue.Factory, factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		Unmarshal(dataValue.Factory, factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	default:
		log.Error("unknown factory type")
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
		log.ErrorFields("", zap.Error(err))
		return
	}
	log.DebugFields("", zap.Reflect("data", dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		dataValue.Factory.ToVal(factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		dataValue.Factory.ToVal(factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	default:
		log.Error("unknown factory type")
	}

	data = `{
		"type": "b",
		"factory": {
			"b": "FactoryB"
		}
	}`
	if err := UnmarshalFromString(data, dataValue); err != nil {
		log.ErrorFields("", zap.Error(err))
		return
	}
	log.DebugFields("", zap.Reflect("data", dataValue))
	switch dataValue.Type {
	case "a":
		factoryValue := &FactoryA{}
		dataValue.Factory.ToVal(factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	case "b":
		factoryValue := &FactoryB{}
		dataValue.Factory.ToVal(factoryValue)
		log.DebugFields("", zap.Reflect("factory", factoryValue))
	default:
		log.Error("unknown factory type")
	}
}
