# encoding: utf-8
"""
This script tests ``GitWildMatchPattern``.
"""
from __future__ import unicode_literals

import sys

try:
	import unittest2 as unittest
except ImportError:
	import unittest

import pathspec.util
from pathspec.patterns.gitwildmatch import GitWildMatchPattern


class GitWildMatchTest(unittest.TestCase):
	"""
	The ``GitWildMatchTest`` class tests the ``GitWildMatchPattern``
	implementation.
	"""

	def test_00_empty(self):
		"""
		Tests an empty pattern.
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('')
		self.assertIsNone(include)
		self.assertIsNone(regex)

	def test_01_absolute_root(self):
		"""
		Tests a single root absolute path pattern.

		This should NOT match any file (according to git check-ignore
		(v2.4.1)).
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('/')
		self.assertIsNone(include)
		self.assertIsNone(regex)

	def test_01_absolute(self):
		"""
		Tests an absolute path pattern.

		This should match:

			an/absolute/file/path
			an/absolute/file/path/foo

		This should NOT match:

			foo/an/absolute/file/path
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('/an/absolute/file/path')
		self.assertTrue(include)
		self.assertEqual(regex, '^an/absolute/file/path(?:/.*)?$')

	def test_01_relative(self):
		"""
		Tests a relative path pattern.

		This should match:

			spam
			spam/
			foo/spam
			spam/foo
			foo/spam/bar
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('spam')
		self.assertTrue(include)
		self.assertEqual(regex, '^(?:.+/)?spam(?:/.*)?$')

	def test_01_relative_nested(self):
		"""
		Tests a relative nested path pattern.

		This should match:

			foo/spam
			foo/spam/bar

		This should **not** match (according to git check-ignore (v2.4.1)):

			bar/foo/spam
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('foo/spam')
		self.assertTrue(include)
		self.assertEqual(regex, '^foo/spam(?:/.*)?$')

	def test_02_comment(self):
		"""
		Tests a comment pattern.
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('# Cork soakers.')
		self.assertIsNone(include)
		self.assertIsNone(regex)

	def test_02_ignore(self):
		"""
		Tests an exclude pattern.

		This should NOT match (according to git check-ignore (v2.4.1)):

			temp/foo
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('!temp')
		self.assertIsNotNone(include)
		self.assertFalse(include)
		self.assertEqual(regex, '^(?:.+/)?temp$')

	def test_03_child_double_asterisk(self):
		"""
		Tests a directory name with a double-asterisk child
		directory.

		This should match:

			spam/bar

		This should **not** match (according to git check-ignore (v2.4.1)):

			foo/spam/bar
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('spam/**')
		self.assertTrue(include)
		self.assertEqual(regex, '^spam/.*$')

	def test_03_inner_double_asterisk(self):
		"""
		Tests a path with an inner double-asterisk directory.

		This should match:

			left/bar/right
			left/foo/bar/right
			left/bar/right/foo

		This should **not** match (according to git check-ignore (v2.4.1)):

			foo/left/bar/right
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('left/**/right')
		self.assertTrue(include)
		self.assertEqual(regex, '^left(?:/.+)?/right(?:/.*)?$')

	def test_03_only_double_asterisk(self):
		"""
		Tests a double-asterisk pattern which matches everything.
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('**')
		self.assertTrue(include)
		self.assertEqual(regex, '^.+$')

	def test_03_parent_double_asterisk(self):
		"""
		Tests a file name with a double-asterisk parent directory.

		This should match:

			foo/spam
			foo/spam/bar
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('**/spam')
		self.assertTrue(include)
		self.assertEqual(regex, '^(?:.+/)?spam(?:/.*)?$')

	def test_04_infix_wildcard(self):
		"""
		Tests a pattern with an infix wildcard.

		This should match:

			foo--bar
			foo-hello-bar
			a/foo-hello-bar
			foo-hello-bar/b
			a/foo-hello-bar/b
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('foo-*-bar')
		self.assertTrue(include)
		self.assertEqual(regex, '^(?:.+/)?foo\\-[^/]*\\-bar(?:/.*)?$')

	def test_04_postfix_wildcard(self):
		"""
		Tests a pattern with a postfix wildcard.

		This should match:

			~temp-
			~temp-foo
			~temp-foo/bar
			foo/~temp-bar
			foo/~temp-bar/baz
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('~temp-*')
		self.assertTrue(include)
		self.assertEqual(regex, '^(?:.+/)?\\~temp\\-[^/]*(?:/.*)?$')

	def test_04_prefix_wildcard(self):
		"""
		Tests a pattern with a prefix wildcard.

		This should match:

			bar.py
			bar.py/
			foo/bar.py
			foo/bar.py/baz
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('*.py')
		self.assertTrue(include)
		self.assertEqual(regex, '^(?:.+/)?[^/]*\\.py(?:/.*)?$')

	def test_05_directory(self):
		"""
		Tests a directory pattern.

		This should match:

			dir/
			foo/dir/
			foo/dir/bar

		This should **not** match:

			dir
		"""
		regex, include = GitWildMatchPattern.pattern_to_regex('dir/')
		self.assertTrue(include)
		self.assertEqual(regex, '^(?:.+/)?dir/.*$')

	def test_06_registered(self):
		"""
		Tests that the pattern is registered.
		"""
		self.assertIs(pathspec.util.lookup_pattern('gitwildmatch'), GitWildMatchPattern)

	def test_06_access_deprecated(self):
		"""
		Tests that the pattern is accessible from the root module using the
		deprecated alias.
		"""
		self.assertTrue(hasattr(pathspec, 'GitIgnorePattern'))
		self.assertTrue(issubclass(pathspec.GitIgnorePattern, GitWildMatchPattern))

	def test_06_registered_deprecated(self):
		"""
		Tests that the pattern is registered under the deprecated alias.
		"""
		self.assertIs(pathspec.util.lookup_pattern('gitignore'), pathspec.GitIgnorePattern)

	def test_07_match_bytes_and_bytes(self):
		"""
		Test byte string patterns matching byte string paths.
		"""
		pattern = GitWildMatchPattern(b'*.py')
		results = set(pattern.match([b'a.py']))
		self.assertEqual(results, set([b'a.py']))

	@unittest.skipIf(sys.version_info[0] >= 3, "Python 3 is strict")
	def test_07_match_bytes_and_unicode(self):
		"""
		Test byte string patterns matching byte string paths.
		"""
		pattern = GitWildMatchPattern(b'*.py')
		results = set(pattern.match(['a.py']))
		self.assertEqual(results, set(['a.py']))

	@unittest.skipIf(sys.version_info[0] == 2, "Python 2 is lenient")
	def test_07_match_bytes_and_unicode_fail(self):
		"""
		Test byte string patterns matching byte string paths.
		"""
		pattern = GitWildMatchPattern(b'*.py')
		with self.assertRaises(TypeError):
			for _ in pattern.match(['a.py']):
				pass

	@unittest.skipIf(sys.version_info[0] >= 3, "Python 3 is strict")
	def test_07_match_unicode_and_bytes(self):
		"""
		Test unicode patterns with byte paths.
		"""
		pattern = GitWildMatchPattern('*.py')
		results = set(pattern.match([b'a.py']))
		self.assertEqual(results, set([b'a.py']))

	@unittest.skipIf(sys.version_info[0] == 2, "Python 2 is lenient")
	def test_07_match_unicode_and_bytes_fail(self):
		"""
		Test unicode patterns with byte paths.
		"""
		pattern = GitWildMatchPattern('*.py')
		with self.assertRaises(TypeError):
			for _ in pattern.match([b'a.py']):
				pass

	def test_07_match_unicode_and_unicode(self):
		"""
		Test unicode patterns with unicode paths.
		"""
		pattern = GitWildMatchPattern('*.py')
		results = set(pattern.match(['a.py']))
		self.assertEqual(results, set(['a.py']))
