// Copyright 2023 mixcode@github
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cookiejar

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

// run all the original tests with marshalling
func TestMarshal(t *testing.T) {
	testMarshalS(t, basicsTests[:])

	testMarshalC(t, updateAndDeleteTests[:])

	testMarshalS(t, chromiumBasicsTests[:])

	testMarshalC(t, chromiumDomainTests[:])

	testMarshalC(t, chromiumDeletionTests[:])
	testMarshalS(t, domainHandlingTests[:])

	// TestExpiration()
	testMarshalS(t, []jarTest{{
		"Expiration.",
		"http://www.host.test",
		[]string{
			"a=1",
			"b=2; max-age=3",
			"c=3; " + expiresIn(3),
			"d=4; max-age=5",
			"e=5; " + expiresIn(5),
			"f=6; max-age=100",
		},
		"a=1 b=2 c=3 d=4 e=5 f=6", // executed at t0 + 1001 ms
		[]query{
			{"http://www.host.test", "a=1 b=2 c=3 d=4 e=5 f=6"}, // t0 + 2002 ms
			{"http://www.host.test", "a=1 d=4 e=5 f=6"},         // t0 + 3003 ms
			{"http://www.host.test", "a=1 d=4 e=5 f=6"},         // t0 + 4004 ms
			{"http://www.host.test", "a=1 f=6"},                 // t0 + 5005 ms
			{"http://www.host.test", "a=1 f=6"},                 // t0 + 6006 ms
		},
	}})
}

// single-entry test
func testMarshalS(t *testing.T, cases []jarTest) {
	for _, test := range cases {
		// prepare a test case
		jar := newTestJar()
		test.prepare(t, jar)

		// save and load to a new jar
		m, err := jar.MarshalJson(false)
		if err != nil {
			t.Fatal(err)
		}
		jar2 := newTestJar()
		err = jar2.MergeJson(m)
		if err != nil {
			t.Fatal(err)
		}

		// execute the test
		test.execute(t, jar2)
	}
}

// continuous, multiple entry test
func testMarshalC(t *testing.T, cases []jarTest) {
	jar := newTestJar()
	for _, test := range cases {
		// prepare a test case
		test.prepare(t, jar)

		// save and load to a new jar
		m, err := jar.MarshalJson(false)
		if err != nil {
			t.Fatal(err)
		}
		jar2 := newTestJar()
		err = jar2.MergeJson(m)
		if err != nil {
			t.Fatal(err)
		}

		// execute tests
		test.execute(t, jar) // also run the master test to update the jar
		test.execute(t, jar2)
	}
}

// prepare() / execute() is separated parts of run() to test marshal/unmarshal
func (test jarTest) prepare(t *testing.T, jar *Jar) {
	now := tNow

	// Populate jar with cookies.
	setCookies := make([]*http.Cookie, len(test.setCookies))
	for i, cs := range test.setCookies {
		cookies := (&http.Response{Header: http.Header{"Set-Cookie": {cs}}}).Cookies()
		if len(cookies) != 1 {
			panic(fmt.Sprintf("Wrong cookie line %q: %#v", cs, cookies))
		}
		setCookies[i] = cookies[0]
	}
	jar.setCookies(mustParseURL(test.fromURL), setCookies, now)
}

func (test jarTest) execute(t *testing.T, jar *Jar) (ok bool) {
	now := tNow

	ok = true

	now = now.Add(1001 * time.Millisecond)

	// Serialize non-expired entries in the form "name1=val1 name2=val2".
	var cs []string
	for _, submap := range jar.entries {
		for _, cookie := range submap {
			if !cookie.Expires.After(now) {
				continue
			}
			cs = append(cs, cookie.Name+"="+cookie.Value)
		}
	}
	sort.Strings(cs)
	got := strings.Join(cs, " ")

	// Make sure jar content matches our expectations.
	if got != test.content {
		t.Errorf("Test %q Content\ngot  %q\nwant %q",
			test.description, got, test.content)
		ok = false
	}

	// Test different calls to Cookies.
	for i, query := range test.queries {
		now = now.Add(1001 * time.Millisecond)
		var s []string
		for _, c := range jar.cookies(mustParseURL(query.toURL), now) {
			s = append(s, c.Name+"="+c.Value)
		}
		if got := strings.Join(s, " "); got != query.want {
			t.Errorf("Test %q #%d\ngot  %q\nwant %q", test.description, i, got, query.want)
			ok = false
		}
	}
	return
}
