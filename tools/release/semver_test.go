package main

import "testing"

func TestParseVersion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   string
		want Version
	}{
		{"v0.0.0", Version{}},
		{"v2.1.0", Version{Major: 2, Minor: 1}},
		{"v1.2.0-beta.1", Version{Major: 1, Minor: 2, Pre: "beta.1"}},
		{"v10.20.30", Version{Major: 10, Minor: 20, Patch: 30}},
	} {
		got, err := ParseVersion(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("ParseVersion(%q) = %+v, %v; want %+v", tc.in, got, err, tc.want)
		}
		if got.String() != tc.in {
			t.Errorf("ParseVersion(%q).String() = %q", tc.in, got)
		}
	}
	for _, in := range []string{"", "1.0.4", "v1.0", "v1.0.0.0", "v01.0.0", "v1.0.0+meta", "v1.0.0-", "v1.x.0", "v-1.0.0"} {
		if v, err := ParseVersion(in); err == nil {
			t.Errorf("ParseVersion(%q) = %v, want error", in, v)
		}
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()
	ordered := []string{"v1.9.9", "v2.0.0-beta.1", "v2.0.0", "v2.0.4", "v2.1.0", "v2.10.0", "v10.0.0"}
	for i := range ordered {
		for j := range ordered {
			a, _ := ParseVersion(ordered[i])
			b, _ := ParseVersion(ordered[j])
			got := a.Compare(b)
			if (i < j && got >= 0) || (i > j && got <= 0) || (i == j && got != 0) {
				t.Errorf("Compare(%s, %s) = %d", a, b, got)
			}
		}
	}
}

func TestBump(t *testing.T) {
	t.Parallel()
	v := Version{Major: 2, Minor: 1, Patch: 3}
	for b, want := range map[Bump]string{BumpNone: "v2.1.3", BumpPatch: "v2.1.4", BumpMinor: "v2.2.0", BumpMajor: "v3.0.0"} {
		if got := v.Bump(b).String(); got != want {
			t.Errorf("%s bump of %s = %s, want %s", b, v, got, want)
		}
	}
}
