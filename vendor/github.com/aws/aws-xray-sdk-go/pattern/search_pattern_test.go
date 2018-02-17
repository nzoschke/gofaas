// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package pattern

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidArgs(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("", "whatever"))
}

func TestMatchExactPositive(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("foo", "foo"))
}

func TestMatchExactNegative(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("foo", "bar"))
}

func TestSignleWildcardPositive(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("fo?", "foo"))
}

func TestSingleWildcardNegative(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("f?o", "boo"))
}

func TestMultipleWildcardPositive(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("?o?", "foo"))
}

func TestMultipleWildcardNegative(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("f??", "boo"))
}

func TestGlobPositive(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("*oo", "foo"))
}

func TestGlobPositiveZeroOrMore(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("foo*", "foo"))
}

func TestGlobNegativeZeroOrMore(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("foo*", "fo0"))
}

func TestGlobNegative(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("fo*", "boo"))
}

func TestGlobAndSinglePositive(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("*o?", "foo"))
}

func TestGlobAndSingleNegative(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("f?*", "boo"))
}

func TestPureWildcard(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("*", "boo"))
}

func TestMisc(t *testing.T) {
	animal1 := "?at"
	animal2 := "?o?se"
	animal3 := "*s"

	vehicle1 := "J*"
	vehicle2 := "????"

	assert.True(t, WildcardMatchCaseInsensitive(animal1, "bat"))
	assert.True(t, WildcardMatchCaseInsensitive(animal1, "cat"))
	assert.True(t, WildcardMatchCaseInsensitive(animal2, "horse"))
	assert.True(t, WildcardMatchCaseInsensitive(animal2, "mouse"))
	assert.True(t, WildcardMatchCaseInsensitive(animal3, "dogs"))
	assert.True(t, WildcardMatchCaseInsensitive(animal3, "horses"))

	assert.True(t, WildcardMatchCaseInsensitive(vehicle1, "Jeep"))
	assert.True(t, WildcardMatchCaseInsensitive(vehicle2, "ford"))
	assert.False(t, WildcardMatchCaseInsensitive(vehicle2, "chevy"))
	assert.True(t, WildcardMatchCaseInsensitive("*", "cAr"))

	assert.True(t, WildcardMatchCaseInsensitive("*/foo", "/bar/foo"))
}

func TestCaseInsensitivity(t *testing.T) {
	assert.True(t, WildcardMatch("Foo", "Foo", false))
	assert.True(t, WildcardMatch("Foo", "Foo", true))

	assert.False(t, WildcardMatch("Foo", "FOO", false))
	assert.True(t, WildcardMatch("Foo", "FOO", true))

	assert.True(t, WildcardMatch("Fo*", "Foo0", false))
	assert.True(t, WildcardMatch("Fo*", "Foo0", true))

	assert.False(t, WildcardMatch("Fo*", "FOo0", false))
	assert.True(t, WildcardMatch("Fo*", "FOO0", true))

	assert.True(t, WildcardMatch("Fo?", "Foo", false))
	assert.True(t, WildcardMatch("Fo?", "Foo", true))

	assert.False(t, WildcardMatch("Fo?", "FOo", false))
	assert.True(t, WildcardMatch("Fo?", "FoO", false))
	assert.True(t, WildcardMatch("Fo?", "FOO", true))
}

func TestLongStrings(t *testing.T) {
	chars := []byte{'a', 'b', 'c', 'd'}
	text := bytes.NewBufferString("a")
	for i := 0; i < 8192; i++ {
		text.WriteString(string(chars[rand.Intn(len(chars))]))
	}
	text.WriteString("b")

	assert.True(t, WildcardMatchCaseInsensitive("a*b", text.String()))
}

func TestNoGlobs(t *testing.T) {
	assert.False(t, WildcardMatchCaseInsensitive("abcd", "abc"))
}

func TestEdgeCaseGlobs(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("", ""))
	assert.True(t, WildcardMatchCaseInsensitive("a", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("*a", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("*a", "ba"))
	assert.True(t, WildcardMatchCaseInsensitive("a*", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("a*", "ab"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a", "aa"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a", "aba"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a", "aaa"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a*", "aa"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a*", "aba"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a*", "aaa"))
	assert.True(t, WildcardMatchCaseInsensitive("a*a*", "aaaaaaaaaaaaaaaaaaaaaaa"))
	assert.True(t, WildcardMatchCaseInsensitive("a*b*a*b*a*b*a*b*a*",
		"akljd9gsdfbkjhaabajkhbbyiaahkjbjhbuykjakjhabkjhbabjhkaabbabbaaakljdfsjklababkjbsdabab"))
	assert.False(t, WildcardMatchCaseInsensitive("a*na*ha", "anananahahanahana"))
}

func TestMultiGlobs(t *testing.T) {
	assert.True(t, WildcardMatchCaseInsensitive("*a", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("**a", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("***a", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("**a*", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("**a**", "a"))

	assert.True(t, WildcardMatchCaseInsensitive("a**b", "ab"))
	assert.True(t, WildcardMatchCaseInsensitive("a**b", "abb"))

	assert.True(t, WildcardMatchCaseInsensitive("*?", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("*?", "aa"))
	assert.True(t, WildcardMatchCaseInsensitive("*??", "aa"))
	assert.False(t, WildcardMatchCaseInsensitive("*???", "aa"))
	assert.True(t, WildcardMatchCaseInsensitive("*?", "aaa"))

	assert.True(t, WildcardMatchCaseInsensitive("?", "a"))
	assert.False(t, WildcardMatchCaseInsensitive("??", "a"))

	assert.True(t, WildcardMatchCaseInsensitive("?*", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("*?", "a"))
	assert.False(t, WildcardMatchCaseInsensitive("?*?", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("?*?", "aa"))
	assert.True(t, WildcardMatchCaseInsensitive("*?*", "a"))

	assert.False(t, WildcardMatchCaseInsensitive("*?*a", "a"))
	assert.True(t, WildcardMatchCaseInsensitive("*?*a*", "ba"))

}
