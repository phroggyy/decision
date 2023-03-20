package metadata

import "encoding/json"

type PrivateMetadata struct {
	ChannelID string      `json:"channel_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func ForChannel(channelID string) *PrivateMetadata {
	return &PrivateMetadata{
		ChannelID: channelID,
	}
}

func Parse(data string) (*PrivateMetadata, error) {
	var m PrivateMetadata

	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func MustParse(data string) *PrivateMetadata {
	m, err := Parse(data)
	if err != nil {
		panic(err)
	}

	return m
}

func (m *PrivateMetadata) WithData(data interface{}) *PrivateMetadata {
	m.Data = data

	return m
}

func (m *PrivateMetadata) String() string {
	v, _ := json.Marshal(m)

	return string(v)
}
