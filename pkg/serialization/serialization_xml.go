package serialization

import (
	"encoding/xml"
)

func init() {
	RegisterSerializer(SerializationTypeXML, &XMLSerialization{})
}

type XMLSerialization struct{}

func (*XMLSerialization) Unmarshal(in []byte, body interface{}) error {
	return xml.Unmarshal(in, body)
}

func (*XMLSerialization) Marshal(body interface{}) ([]byte, error) {
	return xml.Marshal(body)
}
