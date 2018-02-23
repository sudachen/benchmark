package ppftool

import (
	"encoding/json"
	"strconv"
)

func (p0 *Report) FromMap(m map[string]interface{}) error {
	if err := p0.Unit.FromString(m["unit"].(string)); err != nil {
		return err
	}
	p0.Label = m["label"].(string)
	p0.Image = m["image"].(string)
	if u, ok := m["rows"]; ok {
		a := u.([]interface{})
		p0.Rows = make(Rows, 0, len(a))
		for _, z := range a {
			var err error
			w := z.(map[string]interface{})
			r := &Row{}
			if r.Flat, err = strconv.ParseFloat(w["flat"].(string), 64); err != nil {
				return err
			}
			if r.FlatPercent, err = strconv.ParseFloat(w["flat%"].(string), 64); err != nil {
				return err
			}
			if r.SumPercent, err = strconv.ParseFloat(w["sum%"].(string), 64); err != nil {
				return err
			}
			if r.Cum, err = strconv.ParseFloat(w["cum"].(string), 64); err != nil {
				return err
			}
			if r.CumPercent, err = strconv.ParseFloat(w["cum%"].(string), 64); err != nil {
				return err
			}
			r.Function = w["function"].(string)
			p0.Rows = append(p0.Rows, r)
		}
	}
	if u, ok := m["errors"]; ok {
		a := u.([]interface{})
		p0.Errors = make([]string, 0, len(a))
		for _, w := range a {
			p0.Errors = append(p0.Errors, w.(string))
		}
	}
	return nil
}

func (p *Report) UnmarshalJSON(bs []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(bs, &m); err != nil {
		return err
	}
	return p.FromMap(m)
}

func (p *Report) ToMap() map[string]interface{} {
	v := make(map[string]interface{})
	v["label"] = p.Label
	v["image"] = p.Image
	v["unit"] = p.Unit.String()
	r := make([]interface{}, len(p.Rows))
	for i, x := range p.Rows {
		r0 := make(map[string]string)
		r0["flat"] = strconv.FormatFloat(x.Flat, 'f', -1, 64)
		r0["flat%"] = strconv.FormatFloat(x.FlatPercent, 'f', -1, 64)
		r0["cum"] = strconv.FormatFloat(x.Cum, 'f', -1, 64)
		r0["cum%"] = strconv.FormatFloat(x.CumPercent, 'f', -1, 64)
		r0["sum%"] = strconv.FormatFloat(x.SumPercent, 'f', -1, 64)
		r0["function"] = x.Function
		r[i] = r0
	}
	v["rows"] = r
	r = make([]interface{}, len(p.Errors))
	for i, x := range p.Errors {
		r[i] = x
	}
	v["errors"] = r
	return v
}

func (p *Report) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToMap())
}
