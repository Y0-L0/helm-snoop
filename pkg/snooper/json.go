package snooper

import (
	"encoding/json"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type ResultJSON struct {
	ChartName  string          `json:"chartName"`
	Referenced vpath.PathsJSON `json:"referenced,omitempty"`
	Unused     vpath.PathsJSON `json:"unused"`
	Undefined  vpath.PathsJSON `json:"undefined"`
}

func (cs Charts) ToJSON(w io.Writer, showReferenced bool) error {
	var out []ResultJSON
	for _, c := range cs {
		rj := ResultJSON{
			ChartName: c.Name,
		}
		if c.Result != nil {
			rj.Unused = c.Result.Unused.ToJSON()
			rj.Undefined = c.Result.Undefined.ToJSON()
			if showReferenced {
				rj.Referenced = c.Result.Referenced.ToJSON()
			}
		}
		out = append(out, rj)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

// toJSON for backward compatibility with tests.
func (c *Chart) toJSON() ResultJSON {
	rj := ResultJSON{
		ChartName: c.Name,
	}
	if c.Result != nil {
		rj.Referenced = c.Result.Referenced.ToJSON()
		rj.Unused = c.Result.Unused.ToJSON()
		rj.Undefined = c.Result.Undefined.ToJSON()
	}
	return rj
}
