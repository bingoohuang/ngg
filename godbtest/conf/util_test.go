package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindOption_Find(t *testing.T) {
	f := &FindOption{
		Open:  '[',
		Close: ']',
		Quote: '\'',
	}

	// 蚂蚁搬家式的更新大表，批次大小100，每10s或者更新1000条，打印日志
	// %ants -N 100 -log 10s,1000r
	// %ants -query='select ID from zz where MODIFIED < '2023-05-17' [where ID >= :ID] order by ID limit :N'
	// %ants -update='update zz set MODIFIED = now() where ID in (:ID{N})'

	tags := f.FindTags(`select ID from zz where MODIFIED < '2023-05-17' [where ID >= :ID] order by ID limit :N`)
	assert.Equal(t, []*Pos{{From: 48, To: 65}}, tags)

	names := f.FindNamed(`select ID from zz where MODIFIED < '2023-05-17' where ID >= :ID order by ID limit :N`)
	assert.Equal(t, []*Pos{{From: 60, To: 63}, {From: 82, To: 84}}, names)

	names2 := f.FindNamed(`update zz set MODIFIED = now() where ID in (:ID{5})`)
	assert.Equal(t, []*Pos{{From: 44, To: 50, ArgOpen: 47}}, names2)
}
