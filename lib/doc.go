package moon

import (
	"encoding/json"
)

type Doc struct {
	items map[string]interface{}
}

func (d *Doc) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.items)
}
