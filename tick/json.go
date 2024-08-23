package tick

import (
	"encoding/json"
	"time"
)

type Dur struct {
	time.Duration
}

func MakeDur(d time.Duration) Dur {
	return Dur{Duration: d}
}

func (d Dur) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Dur) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' {
		sd := string(b[1 : len(b)-1])
		d.Duration, err = time.ParseDuration(sd)
		return
	}

	var id int64
	id, err = json.Number(string(b)).Int64()
	d.Duration = time.Duration(id)
	return
}
