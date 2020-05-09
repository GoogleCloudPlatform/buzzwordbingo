package main

import "testing"

func TestGetUsername(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"accounts.google.com:example@gmail.com", "example@gmail.com"},
		{"", ""},
	}
	for _, c := range cases {
		got := getEmailFromString(c.in)
		if got != c.want {
			t.Errorf("getEmailFromString(%s)  got %s, want %s", c.in, got, c.want)
		}

	}

}
