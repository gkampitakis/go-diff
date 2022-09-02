// Copyright (c) 2012-2016 The go-diff authors. All rights reserved.
// https://github.com/sergi/go-diff
// See the included LICENSE file for license details.
//
// go-diff is a Go implementation of Google's Diff, Match, and Patch library
// Original library is Copyright (c) 2006 Google Inc.
// http://code.google.com/p/google-diff-match-patch/

package diffmatchpatch

import (
	"fmt"
	"strings"
	"testing"
)

func TestPatchString(t *testing.T) {
	type TestCase struct {
		Patch Patch

		Expected string
	}

	for i, tc := range []TestCase{
		{
			Patch: Patch{
				Start1:  20,
				Start2:  21,
				Length1: 18,
				Length2: 17,

				diffs: []Diff{
					{DiffEqual, "jump"},
					{DiffDelete, "s"},
					{DiffInsert, "ed"},
					{DiffEqual, " over "},
					{DiffDelete, "the"},
					{DiffInsert, "a"},
					{DiffEqual, "\nlaz"},
				},
			},

			Expected: "@@ -21,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n %0Alaz\n",
		},
	} {
		actual := tc.Patch.String()
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %#v", i, tc))
	}
}

func TestPatchFromText(t *testing.T) {
	type TestCase struct {
		Patch string

		ErrorMessagePrefix string
	}

	dmp := New()

	for i, tc := range []TestCase{
		{Patch: "", ErrorMessagePrefix: ""},
		{Patch: "@@ -21,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n %0Alaz\n", ErrorMessagePrefix: ""},
		{Patch: "@@ -1 +1 @@\n-a\n+b\n", ErrorMessagePrefix: ""},
		{Patch: "@@ -1,3 +0,0 @@\n-abc\n", ErrorMessagePrefix: ""},
		{Patch: "@@ -0,0 +1,3 @@\n+abc\n", ErrorMessagePrefix: ""},
		{Patch: "@@ _0,0 +0,0 @@\n+abc\n", ErrorMessagePrefix: "Invalid patch string: @@ _0,0 +0,0 @@"},
		{Patch: "Bad\nPatch\n", ErrorMessagePrefix: "Invalid patch string"},
	} {
		patches, err := dmp.PatchFromText(tc.Patch)
		if tc.ErrorMessagePrefix == "" {
			assertEqual(t, nil, err)

			if tc.Patch == "" {
				assertEqual(t, []Patch{}, patches, fmt.Sprintf("Test case #%d, %#v", i, tc))
			} else {
				assertEqual(t, tc.Patch, patches[0].String(), fmt.Sprintf("Test case #%d, %#v", i, tc))
			}
		} else {
			e := err.Error()
			if strings.HasPrefix(e, tc.ErrorMessagePrefix) {
				e = tc.ErrorMessagePrefix
			}
			assertEqual(t, tc.ErrorMessagePrefix, e)
		}
	}

	diffs := []Diff{
		{DiffDelete, "`1234567890-=[]\\;',./"},
		{DiffInsert, "~!@#$%^&*()_+{}|:\"<>?"},
	}

	patches, err := dmp.PatchFromText(
		"@@ -1,21 +1,21 @@\n-%601234567890-=%5B%5D%5C;',./\n+~!@#$%25%5E&*()_+%7B%7D%7C:%22%3C%3E?\n",
	)
	assertEqual(t, 1, len(patches))
	assertEqual(t, diffs,
		patches[0].diffs,
	)
	assertEqual(t, nil, err)
}

func TestPatchToText(t *testing.T) {
	type TestCase struct {
		Patch string
	}

	dmp := New()

	for i, tc := range []TestCase{
		{"@@ -21,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n"},
		{"@@ -1,9 +1,9 @@\n-f\n+F\n oo+fooba\n@@ -7,9 +7,9 @@\n obar\n-,\n+.\n  tes\n"},
	} {
		patches, err := dmp.PatchFromText(tc.Patch)
		assertEqual(t, nil, err)

		actual := dmp.PatchToText(patches)
		assertEqual(t, tc.Patch, actual, fmt.Sprintf("Test case #%d, %#v", i, tc))
	}
}

func TestPatchAddContext(t *testing.T) {
	type TestCase struct {
		Name string

		Patch string
		Text  string

		Expected string
	}

	dmp := New()
	dmp.PatchMargin = 4

	for i, tc := range []TestCase{
		{
			Name:     "Simple case",
			Patch:    "@@ -21,4 +21,10 @@\n-jump\n+somersault\n",
			Text:     "The quick brown fox jumps over the lazy dog.",
			Expected: "@@ -17,12 +17,18 @@\n fox \n-jump\n+somersault\n s ov\n",
		},
		{
			Name:     "Not enough trailing context",
			Patch:    "@@ -21,4 +21,10 @@\n-jump\n+somersault\n",
			Text:     "The quick brown fox jumps.",
			Expected: "@@ -17,10 +17,16 @@\n fox \n-jump\n+somersault\n s.\n",
		},
		{
			Name:     "Not enough leading context",
			Patch:    "@@ -3 +3,2 @@\n-e\n+at\n",
			Text:     "The quick brown fox jumps.",
			Expected: "@@ -1,7 +1,8 @@\n Th\n-e\n+at\n  qui\n",
		},
		{
			Name:     "Ambiguity",
			Patch:    "@@ -3 +3,2 @@\n-e\n+at\n",
			Text:     "The quick brown fox jumps.  The quick brown fox crashes.",
			Expected: "@@ -1,27 +1,28 @@\n Th\n-e\n+at\n  quick brown fox jumps. \n",
		},
	} {
		patches, err := dmp.PatchFromText(tc.Patch)
		assertEqual(t, nil, err)

		actual := dmp.PatchAddContext(patches[0], tc.Text)
		assertEqual(t, tc.Expected, actual.String(), fmt.Sprintf("Test case #%d, %s", i, tc.Name))
	}
}

func TestPatchMakeAndPatchToText(t *testing.T) {
	type TestCase struct {
		Name string

		Input1 interface{}
		Input2 interface{}
		Input3 interface{}

		Expected string
	}

	dmp := New()

	text1 := "The quick brown fox jumps over the lazy dog."
	text2 := "That quick brown fox jumped over a lazy dog."

	for i, tc := range []TestCase{
		{
			Name:     "Null case",
			Input1:   "",
			Input2:   "",
			Input3:   nil,
			Expected: "",
		},
		{
			Name:     "Text2+Text1 inputs",
			Input1:   text2,
			Input2:   text1,
			Input3:   nil,
			Expected: "@@ -1,8 +1,7 @@\n Th\n-at\n+e\n  qui\n@@ -21,17 +21,18 @@\n jump\n-ed\n+s\n  over \n-a\n+the\n  laz\n",
		},
		{
			Name:     "Text1+Text2 inputs",
			Input1:   text1,
			Input2:   text2,
			Input3:   nil,
			Expected: "@@ -1,11 +1,12 @@\n Th\n-e\n+at\n  quick b\n@@ -22,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n",
		},
		{
			Name:     "Diff input",
			Input1:   dmp.DiffMain(text1, text2, false),
			Input2:   nil,
			Input3:   nil,
			Expected: "@@ -1,11 +1,12 @@\n Th\n-e\n+at\n  quick b\n@@ -22,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n",
		},
		{
			Name:     "Text1+Diff inputs",
			Input1:   text1,
			Input2:   dmp.DiffMain(text1, text2, false),
			Input3:   nil,
			Expected: "@@ -1,11 +1,12 @@\n Th\n-e\n+at\n  quick b\n@@ -22,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n",
		},
		{
			Name:     "Text1+Text2+Diff inputs (deprecated)",
			Input1:   text1,
			Input2:   text2,
			Input3:   dmp.DiffMain(text1, text2, false),
			Expected: "@@ -1,11 +1,12 @@\n Th\n-e\n+at\n  quick b\n@@ -22,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n",
		},
		{
			Name:     "Character encoding",
			Input1:   "`1234567890-=[]\\;',./",
			Input2:   "~!@#$%^&*()_+{}|:\"<>?",
			Input3:   nil,
			Expected: "@@ -1,21 +1,21 @@\n-%601234567890-=%5B%5D%5C;',./\n+~!@#$%25%5E&*()_+%7B%7D%7C:%22%3C%3E?\n",
		},
		{
			Name:     "Long string with repeats",
			Input1:   strings.Repeat("abcdef", 100),
			Input2:   strings.Repeat("abcdef", 100) + "123",
			Input3:   nil,
			Expected: "@@ -573,28 +573,31 @@\n cdefabcdefabcdefabcdefabcdef\n+123\n",
		},
		{
			Name:     "Corner case of #31 fixed by #32",
			Input1:   "2016-09-01T03:07:14.807830741Z",
			Input2:   "2016-09-01T03:07:15.154800781Z",
			Input3:   nil,
			Expected: "@@ -15,16 +15,16 @@\n 07:1\n+5.15\n 4\n-.\n 80\n+0\n 78\n-3074\n 1Z\n",
		},
	} {
		var patches []Patch
		if tc.Input3 != nil {
			patches = dmp.PatchMake(tc.Input1, tc.Input2, tc.Input3)
		} else if tc.Input2 != nil {
			patches = dmp.PatchMake(tc.Input1, tc.Input2)
		} else if ps, ok := tc.Input1.([]Patch); ok {
			patches = ps
		} else {
			patches = dmp.PatchMake(tc.Input1)
		}

		actual := dmp.PatchToText(patches)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %s", i, tc.Name))
	}

	// Corner case of #28 wrong patch with timeout of 0
	dmp.DiffTimeout = 0

	text1 = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ut risus et " +
		"enim consectetur convallis a non ipsum. Sed nec nibh cursus, interdum libero vel."
	text2 = "Lorem a ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ut risus et " +
		"enim consectetur convallis a non ipsum. Sed nec nibh cursus, interdum liberovel."

	diffs := dmp.DiffMain(text1, text2, true)
	// Additional check that the diff texts are equal to the originals even if we are using DiffMain with checklines=true #29
	assertEqual(t, text1, dmp.DiffText1(diffs))
	assertEqual(t, text2, dmp.DiffText2(diffs))

	patches := dmp.PatchMake(text1, diffs)

	actual := dmp.PatchToText(patches)
	assertEqual(
		t,
		"@@ -1,14 +1,16 @@\n Lorem \n+a \n ipsum do\n@@ -148,13 +148,12 @@\n m libero\n- \n vel.\n",
		actual,
	)

	// Check that empty Patch array is returned for no parameter call
	patches = dmp.PatchMake()
	assertEqual(t, []Patch{}, patches)
}

func TestPatchSplitMax(t *testing.T) {
	type TestCase struct {
		Text1 string
		Text2 string

		Expected string
	}

	dmp := New()

	for i, tc := range []TestCase{
		{
			Text1: "abcdefghijklmnopqrstuvwxyz01234567890",
			Text2: "XabXcdXefXghXijXklXmnXopXqrXstXuvXwxXyzX01X23X45X67X89X0",
			Expected: "@@ -1,32 +1,46 @@\n+X\n ab\n+X\n cd\n+X\n ef\n+X\n gh\n+X\n ij\n+X\n kl\n+X\n mn\n+X\n op\n+X\n qr\n+X\n " +
				"st\n+X\n uv\n+X\n wx\n+X\n yz\n+X\n 012345\n@@ -25,13 +39,18 @@\n zX01\n+X\n 23\n+X\n 45\n+X\n 67\n+X\n 89\n+X\n 0\n",
		},
		{
			Text1:    "abcdef1234567890123456789012345678901234567890123456789012345678901234567890uvwxyz",
			Text2:    "abcdefuvwxyz",
			Expected: "@@ -3,78 +3,8 @@\n cdef\n-1234567890123456789012345678901234567890123456789012345678901234567890\n uvwx\n",
		},
		{
			Text1: "1234567890123456789012345678901234567890123456789012345678901234567890",
			Text2: "abc",
			Expected: "@@ -1,32 +1,4 @@\n-1234567890123456789012345678\n 9012\n@@ -29,32 +1,4 @@\n-9012345678901234" +
				"567890123456\n 7890\n@@ -57,14 +1,3 @@\n-78901234567890\n+abc\n",
		},
		{
			Text1: "abcdefghij , h : 0 , t : 1 abcdefghij , h : 0 , t : 1 abcdefghij , h : 0 , t : 1",
			Text2: "abcdefghij , h : 1 , t : 1 abcdefghij , h : 1 , t : 1 abcdefghij , h : 0 , t : 1",
			Expected: "@@ -2,32 +2,32 @@\n bcdefghij , h : \n-0\n+1\n  , t : 1 abcdef\n@@ -29,32 +29,32 @@\n " +
				"bcdefghij , h : \n-0\n+1\n  , t : 1 abcdef\n",
		},
	} {
		patches := dmp.PatchMake(tc.Text1, tc.Text2)
		patches = dmp.PatchSplitMax(patches)

		actual := dmp.PatchToText(patches)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %#v", i, tc))
	}
}

func TestPatchAddPadding(t *testing.T) {
	type TestCase struct {
		Name string

		Text1 string
		Text2 string

		Expected            string
		ExpectedWithPadding string
	}

	dmp := New()

	for i, tc := range []TestCase{
		{
			Name:                "Both edges full",
			Text1:               "",
			Text2:               "test",
			Expected:            "@@ -0,0 +1,4 @@\n+test\n",
			ExpectedWithPadding: "@@ -1,8 +1,12 @@\n %01%02%03%04\n+test\n %01%02%03%04\n",
		},
		{
			Name:                "Both edges partial",
			Text1:               "XY",
			Text2:               "XtestY",
			Expected:            "@@ -1,2 +1,6 @@\n X\n+test\n Y\n",
			ExpectedWithPadding: "@@ -2,8 +2,12 @@\n %02%03%04X\n+test\n Y%01%02%03\n",
		},
		{
			Name:                "Both edges none",
			Text1:               "XXXXYYYY",
			Text2:               "XXXXtestYYYY",
			Expected:            "@@ -1,8 +1,12 @@\n XXXX\n+test\n YYYY\n",
			ExpectedWithPadding: "@@ -5,8 +5,12 @@\n XXXX\n+test\n YYYY\n",
		},
	} {
		patches := dmp.PatchMake(tc.Text1, tc.Text2)

		actual := dmp.PatchToText(patches)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %s", i, tc.Name))

		dmp.PatchAddPadding(patches)

		actualWithPadding := dmp.PatchToText(patches)
		assertEqual(
			t,
			tc.ExpectedWithPadding,
			actualWithPadding,
			fmt.Sprintf("Test case #%d, %s", i, tc.Name),
		)
	}
}

func TestPatchApply(t *testing.T) {
	type TestCase struct {
		Name string

		Text1    string
		Text2    string
		TextBase string

		Expected        string
		ExpectedApplies []bool
	}

	dmp := New()
	dmp.MatchDistance = 1000
	dmp.MatchThreshold = 0.5
	dmp.PatchDeleteThreshold = 0.5

	for i, tc := range []TestCase{
		{
			Name:  "Null case",
			Text1: "",
			Text2: "", TextBase: "Hello world.",
			Expected:        "Hello world.",
			ExpectedApplies: []bool{},
		},
		{
			Name:            "Exact match",
			Text1:           "The quick brown fox jumps over the lazy dog.",
			Text2:           "That quick brown fox jumped over a lazy dog.",
			TextBase:        "The quick brown fox jumps over the lazy dog.",
			Expected:        "That quick brown fox jumped over a lazy dog.",
			ExpectedApplies: []bool{true, true},
		},
		{
			Name:            "Partial match",
			Text1:           "The quick brown fox jumps over the lazy dog.",
			Text2:           "That quick brown fox jumped over a lazy dog.",
			TextBase:        "The quick red rabbit jumps over the tired tiger.",
			Expected:        "That quick red rabbit jumped over a tired tiger.",
			ExpectedApplies: []bool{true, true},
		},
		{
			Name:            "Failed match",
			Text1:           "The quick brown fox jumps over the lazy dog.",
			Text2:           "That quick brown fox jumped over a lazy dog.",
			TextBase:        "I am the very model of a modern major general.",
			Expected:        "I am the very model of a modern major general.",
			ExpectedApplies: []bool{false, false},
		},
		{
			Name:            "Big delete, small Diff",
			Text1:           "x1234567890123456789012345678901234567890123456789012345678901234567890y",
			Text2:           "xabcy",
			TextBase:        "x123456789012345678901234567890-----++++++++++-----123456789012345678901234567890y",
			Expected:        "xabcy",
			ExpectedApplies: []bool{true, true},
		},
		{
			Name:            "Big delete, big Diff 1",
			Text1:           "x1234567890123456789012345678901234567890123456789012345678901234567890y",
			Text2:           "xabcy",
			TextBase:        "x12345678901234567890---------------++++++++++---------------12345678901234567890y",
			Expected:        "xabc12345678901234567890---------------++++++++++---------------12345678901234567890y",
			ExpectedApplies: []bool{false, true},
		},
	} {
		patches := dmp.PatchMake(tc.Text1, tc.Text2)

		actual, actualApplies := dmp.PatchApply(patches, tc.TextBase)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %s", i, tc.Name))
		assertEqual(
			t,
			tc.ExpectedApplies,
			actualApplies,
			fmt.Sprintf("Test case #%d, %s", i, tc.Name),
		)
	}

	dmp.PatchDeleteThreshold = 0.6

	for i, tc := range []TestCase{
		{
			Name:     "Big delete, big Diff 2",
			Text1:    "x1234567890123456789012345678901234567890123456789012345678901234567890y",
			Text2:    "xabcy",
			TextBase: "x12345678901234567890---------------++++++++++---------------12345678901234567890y",
			Expected: "xabcy", ExpectedApplies: []bool{true, true},
		},
	} {
		patches := dmp.PatchMake(tc.Text1, tc.Text2)

		actual, actualApplies := dmp.PatchApply(patches, tc.TextBase)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %s", i, tc.Name))
		assertEqual(
			t,
			tc.ExpectedApplies,
			actualApplies,
			fmt.Sprintf("Test case #%d, %s", i, tc.Name),
		)
	}

	dmp.MatchDistance = 0
	dmp.MatchThreshold = 0.0
	dmp.PatchDeleteThreshold = 0.5

	for i, tc := range []TestCase{
		{
			Name:            "Compensate for failed patch",
			Text1:           "abcdefghijklmnopqrstuvwxyz--------------------1234567890",
			Text2:           "abcXXXXXXXXXXdefghijklmnopqrstuvwxyz--------------------1234567YYYYYYYYYY890",
			TextBase:        "ABCDEFGHIJKLMNOPQRSTUVWXYZ--------------------1234567890",
			Expected:        "ABCDEFGHIJKLMNOPQRSTUVWXYZ--------------------1234567YYYYYYYYYY890",
			ExpectedApplies: []bool{false, true},
		},
	} {
		patches := dmp.PatchMake(tc.Text1, tc.Text2)

		actual, actualApplies := dmp.PatchApply(patches, tc.TextBase)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %s", i, tc.Name))
		assertEqual(
			t,
			tc.ExpectedApplies,
			actualApplies,
			fmt.Sprintf("Test case #%d, %s", i, tc.Name),
		)
	}

	dmp.MatchThreshold = 0.5
	dmp.MatchDistance = 1000

	for i, tc := range []TestCase{
		{
			Name:            "No side effects",
			Text1:           "",
			Text2:           "test",
			TextBase:        "",
			Expected:        "test",
			ExpectedApplies: []bool{true},
		},
		{
			Name:            "No side effects with major delete",
			Text1:           "The quick brown fox jumps over the lazy dog.",
			Text2:           "Woof",
			TextBase:        "The quick brown fox jumps over the lazy dog.",
			Expected:        "Woof",
			ExpectedApplies: []bool{true, true},
		},
		{
			Name:  "Edge exact match",
			Text1: "", Text2: "test",
			TextBase:        "",
			Expected:        "test",
			ExpectedApplies: []bool{true},
		},
		{
			Name:            "Near edge exact match",
			Text1:           "XY",
			Text2:           "XtestY",
			TextBase:        "XY",
			Expected:        "XtestY",
			ExpectedApplies: []bool{true},
		},
		{
			Name:            "Edge partial match",
			Text1:           "y",
			Text2:           "y123",
			TextBase:        "x",
			Expected:        "x123",
			ExpectedApplies: []bool{true},
		},
	} {
		patches := dmp.PatchMake(tc.Text1, tc.Text2)

		actual, actualApplies := dmp.PatchApply(patches, tc.TextBase)
		assertEqual(t, tc.Expected, actual, fmt.Sprintf("Test case #%d, %s", i, tc.Name))
		assertEqual(
			t,
			tc.ExpectedApplies,
			actualApplies,
			fmt.Sprintf("Test case #%d, %s", i, tc.Name),
		)
	}
}
