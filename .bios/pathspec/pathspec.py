# encoding: utf-8
"""
This module provides an object oriented interface for pattern matching
of files.
"""

import collections

from . import util
from .compat import iterkeys, izip_longest, string_types


class PathSpec(object):
	"""
	The :class:`PathSpec` class is a wrapper around a list of compiled
	:class:`.Pattern` instances.
	"""

	def __init__(self, patterns):
		"""
		Initializes the :class:`PathSpec` instance.

		*patterns* (:class:`~collections.Container` or :class:`~collections.Iterable`)
		yields each compiled pattern (:class:`.Pattern`).
		"""

		self.patterns = patterns if isinstance(patterns, collections.Container) else list(patterns)
		"""
		*patterns* (:class:`~collections.Container` of :class:`.Pattern`)
		contains the compiled patterns.
		"""

	def __eq__(self, other):
		"""
		Tests the equality of this path-spec with *other* (:class:`PathSpec`)
		by comparing their :attr:`~PathSpec.patterns` attributes.
		"""
		if isinstance(other, PathSpec):
			paired_patterns = izip_longest(self.patterns, other.patterns)
			return all(a == b for a, b in paired_patterns)
		else:
			return NotImplemented

	def __len__(self):
		"""
		Returns the number of compiled patterns this path-spec contains
		(:class:`int`).
		"""
		return len(self.patterns)

	@classmethod
	def from_lines(cls, pattern_factory, lines):
		"""
		Compiles the pattern lines.

		*pattern_factory* can be either the name of a registered pattern
		factory (:class:`str`), or a :class:`~collections.Callable` used to
		compile patterns. It must accept an uncompiled pattern (:class:`str`)
		and return the compiled pattern (:class:`.Pattern`).

		*lines* (:class:`~collections.Iterable`) yields each uncompiled
		pattern (:class:`str`). This simply has to yield each line so it can
		be a :class:`file` (e.g., from :func:`open` or :class:`io.StringIO`)
		or the result from :meth:`str.splitlines`.

		Returns the :class:`PathSpec` instance.
		"""
		if isinstance(pattern_factory, string_types):
			pattern_factory = util.lookup_pattern(pattern_factory)
		if not callable(pattern_factory):
			raise TypeError("pattern_factory:{0!r} is not callable.".format(pattern_factory))

		lines = [pattern_factory(line) for line in lines if line]
		return cls(lines)

	def match_file(self, file, separators=None):
		"""
		Matches the file to this path-spec.

		*file* (:class:`str`) is the file path to be matched against
		:attr:`self.patterns <PathSpec.patterns>`.

		*separators* (:class:`~collections.Container` of :class:`str`)
		optionally contains the path separators to normalize. See
		:func:`~pathspec.util.normalize_file` for more information.

		Returns :data:`True` if *file* matched; otherwise, :data:`False`.
		"""
		norm_file = util.normalize_file(file, separators=separators)
		return util.match_file(self.patterns, norm_file)

	def match_files(self, files, separators=None):
		"""
		Matches the files to this path-spec.

		*files* (:class:`~collections.Iterable` of :class:`str`) contains
		the file paths to be matched against :attr:`self.patterns
		<PathSpec.patterns>`.

		*separators* (:class:`~collections.Container` of :class:`str`)
		optionally contains the path separators to normalize. See
		:func:`~pathspec.util.normalize_file` for more information.

		Returns the matched files (:class:`~collections.Iterable` of
		:class:`str`).
		"""
		file_map = util.normalize_files(files, separators=separators)
		matched_files = util.match_files(self.patterns, iterkeys(file_map))
		for path in matched_files:
			yield file_map[path]

	def match_tree(self, root):
		"""
		Walks the specified root path for all files and matches them to this
		path-spec.

		*root* (:class:`str`) is the root directory to search for files.

		Returns the matched files (:class:`~collections.Iterable` of
		:class:`str`).
		"""
		files = util.iter_tree(root)
		return self.match_files(files)
