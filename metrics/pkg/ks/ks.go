package ks

type Ks struct {
	Keys [20]string
}

func Kn(n int, s string) *Ks {
	ks := &Ks{}
	ks.Keys[n-1] = s
	return ks
}

func (ks *Ks) Kn(n int, s string) *Ks {
	ks.Keys[n-1] = s
	return ks
}

func K4(s string) *Ks  { return Kn(4, s) }
func K5(s string) *Ks  { return Kn(5, s) }
func K6(s string) *Ks  { return Kn(6, s) }
func K7(s string) *Ks  { return Kn(7, s) }
func K8(s string) *Ks  { return Kn(8, s) }
func K9(s string) *Ks  { return Kn(9, s) }
func K10(s string) *Ks { return Kn(10, s) }
func K11(s string) *Ks { return Kn(11, s) }
func K12(s string) *Ks { return Kn(12, s) }
func K13(s string) *Ks { return Kn(13, s) }
func K14(s string) *Ks { return Kn(14, s) }
func K15(s string) *Ks { return Kn(15, s) }
func K16(s string) *Ks { return Kn(16, s) }
func K17(s string) *Ks { return Kn(17, s) }
func K18(s string) *Ks { return Kn(18, s) }
func K19(s string) *Ks { return Kn(19, s) }
func K20(s string) *Ks { return Kn(20, s) }

func (ks *Ks) K4(s string) *Ks  { return ks.Kn(4, s) }
func (ks *Ks) K5(s string) *Ks  { return ks.Kn(5, s) }
func (ks *Ks) K6(s string) *Ks  { return ks.Kn(6, s) }
func (ks *Ks) K7(s string) *Ks  { return ks.Kn(7, s) }
func (ks *Ks) K8(s string) *Ks  { return ks.Kn(8, s) }
func (ks *Ks) K9(s string) *Ks  { return ks.Kn(9, s) }
func (ks *Ks) K10(s string) *Ks { return ks.Kn(10, s) }
func (ks *Ks) K11(s string) *Ks { return ks.Kn(11, s) }
func (ks *Ks) K12(s string) *Ks { return ks.Kn(12, s) }
func (ks *Ks) K13(s string) *Ks { return ks.Kn(13, s) }
func (ks *Ks) K14(s string) *Ks { return ks.Kn(14, s) }
func (ks *Ks) K15(s string) *Ks { return ks.Kn(15, s) }
func (ks *Ks) K16(s string) *Ks { return ks.Kn(16, s) }
func (ks *Ks) K17(s string) *Ks { return ks.Kn(17, s) }
func (ks *Ks) K18(s string) *Ks { return ks.Kn(18, s) }
func (ks *Ks) K19(s string) *Ks { return ks.Kn(19, s) }
func (ks *Ks) K20(s string) *Ks { return ks.Kn(20, s) }
