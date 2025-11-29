package v1

import "encoding/json"

type jsonResourceSelector struct {
	ApiVersion  string            `json:"apiVersion"`
	Kind        string            `json:"kind"`
	MatchName   *string            `json:"matchName,omitempty"`
	MatchLabels *MatchLabels `json:"matchLabels,omitempty"`
	Namespace   *string           `json:"namespace,omitempty"`
}

func (r *ResourceSelector) UnmarshalJSON(data []byte) error {
    var tmp jsonResourceSelector

    if err := json.Unmarshal(data, &tmp); err != nil {
        return err
    }

    r.ApiVersion = tmp.ApiVersion
    r.Kind = tmp.Kind
    r.Namespace = tmp.Namespace

    switch {
    case tmp.MatchName != nil:
        r.Match = &ResourceSelector_MatchName{
            MatchName: *tmp.MatchName,
        }

    case tmp.MatchLabels != nil:
        r.Match = &ResourceSelector_MatchLabels{
            MatchLabels: tmp.MatchLabels,
        }

    default:
        r.Match = nil
    }

    return nil
}

func (r *ResourceSelector) MarshalJSON() ([]byte, error) {
	var tmp jsonResourceSelector

	tmp.ApiVersion = r.ApiVersion
	tmp.Kind = r.Kind
	tmp.Namespace = r.Namespace

	switch m := r.Match.(type) {
	case *ResourceSelector_MatchName:
		tmp.MatchName = &m.MatchName
	case *ResourceSelector_MatchLabels:
	  tmp.MatchLabels = m.MatchLabels
	}

	return json.Marshal(tmp)
}

