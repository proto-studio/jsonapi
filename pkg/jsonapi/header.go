package jsonapi

type Version string

const (
	Version_1_0 = "1.0"
	Version_1_1 = "1.1"
)

type Extension struct {
	URI    string
	Prefix string
}

type Profile struct {
	URI    string
	Prefix string
}

type Header struct {
	Version Version        `json:"version"`
	Ext     []Extension    `json:"ext"`
	Profile []Profile      `json:"profile"`
	Meta    map[string]any `json:"meta"`
}
