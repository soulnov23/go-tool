package serialization

import (
	"gopkg.in/yaml.v3"
)

func init() {
	RegisterSerializer(SerializationTypeYAML, &YAMLSerialization{})
}

type YAMLSerialization struct{}

func (s *YAMLSerialization) Unmarshal(in []byte, body interface{}) error {
	return yaml.Unmarshal(in, body)
}

func (s *YAMLSerialization) Marshal(body interface{}) ([]byte, error) {
	return yaml.Marshal(body)
}
