package inventory

import "go.yaml.in/yaml/v3"

func unmarshalYAML(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func marshalYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}
