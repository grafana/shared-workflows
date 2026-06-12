package refprotection

import "testing"

func TestRefNameMatches(t *testing.T) {
	cases := []struct {
		name    string
		include []string
		exclude []string
		ref     string
		want    bool
	}{
		{"all matches tag", []string{"~ALL"}, nil, "refs/tags/v1.9.2", true},
		{"glob v* matches", []string{"refs/tags/v*"}, nil, "refs/tags/v0.39.0", true},
		{"glob crosses slash", []string{"refs/tags/*"}, nil, "refs/tags/app/2.9.4", true},
		{"no match", []string{"refs/tags/v*"}, nil, "refs/tags/app-2.9.4", false},
		{"exact literal", []string{"refs/tags/v1.0.0"}, nil, "refs/tags/v1.0.0", true},
		{"excluded wins", []string{"~ALL"}, []string{"refs/tags/v0*"}, "refs/tags/v0.39.0", false},
		{"included not excluded", []string{"~ALL"}, []string{"refs/tags/v0*"}, "refs/tags/v1.0.0", true},
		{"default-branch token never matches tag", []string{"~DEFAULT_BRANCH"}, nil, "refs/tags/v1", false},
		{"empty include", nil, nil, "refs/tags/v1", false},
		{"question mark single char", []string{"refs/tags/v?"}, nil, "refs/tags/v1", true},
		{"question mark not slash-greedy", []string{"refs/tags/v?"}, nil, "refs/tags/v12", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RefNameMatches(c.include, c.exclude, c.ref); got != c.want {
				t.Errorf("RefNameMatches(%v, %v, %q) = %v, want %v",
					c.include, c.exclude, c.ref, got, c.want)
			}
		})
	}
}
