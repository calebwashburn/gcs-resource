package versions_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/frodenas/gcs-resource/versions"
)

type MatchFunc func(paths []string, pattern string) ([]string, error)

var ItMatchesPaths = func(matchFunc MatchFunc) {
	Describe("checking if paths in the bucket should be searched", func() {
		Context("when given an empty list of paths", func() {
			It("returns an empty list of matches", func() {
				result, err := versions.Match([]string{}, "regex")

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})

		Context("when given a single path", func() {
			It("returns it in a singleton list if it matches the regex", func() {
				paths := []string{"abc"}
				regex := "abc"

				result, err := versions.Match(paths, regex)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(ConsistOf("abc"))
			})

			It("returns an empty list if it does not match the regexp", func() {
				paths := []string{"abc"}
				regex := "ad"

				result, err := versions.Match(paths, regex)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			})

			It("accepts full regexes", func() {
				paths := []string{"abc"}
				regex := "a.*c"

				result, err := versions.Match(paths, regex)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(ConsistOf("abc"))
			})

			It("errors when the regex is bad", func() {
				paths := []string{"abc"}
				regex := "a(c"

				_, err := versions.Match(paths, regex)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when given a multiple paths", func() {
			It("returns the matches", func() {
				paths := []string{"abc", "bcd"}
				regex := ".*bc.*"

				result, err := versions.Match(paths, regex)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(ConsistOf("abc", "bcd"))
			})

			It("returns an empty list if none match the regexp", func() {
				paths := []string{"abc", "def"}
				regex := "ge.*h"

				result, err := versions.Match(paths, regex)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})
	})
}

var _ = Describe("Match", func() {
	Describe("Match", func() {
		ItMatchesPaths(versions.Match)

		It("does not contain files that are in some subdirectory that is not explicitly mentioned", func() {
			paths := []string{"folder/abc", "abc"}
			regex := "abc"

			result, err := versions.Match(paths, regex)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(ConsistOf("abc"))
		})
	})

	Describe("MatchUnanchored", func() {
		ItMatchesPaths(versions.MatchUnanchored)
	})
})

var _ = Describe("Prefix", func() {
	It("turns a regexp into a limiter for gcs", func() {
		By("having a directory prefix in the simple case")
		Expect(versions.Prefix("hello/(.*).tgz")).To(Equal("hello/"))
		Expect(versions.Prefix("hello/world-(.*)")).To(Equal("hello/"))
		Expect(versions.Prefix("hello-world/some-file-(.*)")).To(Equal("hello-world/"))

		By("not having a prefix if there is no parent directory")
		Expect(versions.Prefix("(.*).tgz")).To(Equal(""))
		Expect(versions.Prefix("hello-(.*).tgz")).To(Equal(""))

		By("skipping regexp path names")
		Expect(versions.Prefix("hello/(.*)/what.txt")).To(Equal("hello/"))

		By("handling escaped regexp characters")
		Expect(versions.Prefix(`hello/cruel\[\\\^\$\.\|\?\*\+\(\)world/fizz-(.*).tgz`)).To(Equal(`hello/cruel[\^$.|?*+()world/`))

		By("handling regexp-specific escapes")
		Expect(versions.Prefix(`hello/\d{3}/fizz-(.*).tgz`)).To(Equal(`hello/`))
		Expect(versions.Prefix(`hello/\d/fizz-(.*).tgz`)).To(Equal(`hello/`))
	})
})

var _ = Describe("Extract", func() {
	Context("when the path does not contain extractable information", func() {
		It("doesn't extract it", func() {
			result, ok := versions.Extract("abc.tgz", "abc-(.*).tgz")
			Expect(ok).To(BeFalse())
			Expect(result).To(BeZero())
		})
	})

	Context("when the path contains extractable information", func() {
		It("extracts it", func() {
			result, ok := versions.Extract("abc-105.tgz", "abc-(.*).tgz")
			Expect(ok).To(BeTrue())

			Expect(result.Path).To(Equal("abc-105.tgz"))
			Expect(result.Version.String()).To(Equal("105.0.0"))
			Expect(result.VersionNumber).To(Equal("105"))
		})

		It("extracts semantics version numbers", func() {
			result, ok := versions.Extract("abc-1.0.5.tgz", "abc-(.*).tgz")
			Expect(ok).To(BeTrue())

			Expect(result.Path).To(Equal("abc-1.0.5.tgz"))
			Expect(result.Version.String()).To(Equal("1.0.5"))
			Expect(result.VersionNumber).To(Equal("1.0.5"))
		})

		It("takes the first match if there are many", func() {
			result, ok := versions.Extract("abc-1.0.5-def-2.3.4.tgz", "abc-(.*)-def-(.*).tgz")
			Expect(ok).To(BeTrue())

			Expect(result.Path).To(Equal("abc-1.0.5-def-2.3.4.tgz"))
			Expect(result.Version.String()).To(Equal("1.0.5"))
			Expect(result.VersionNumber).To(Equal("1.0.5"))
		})

		It("extracts a named group called 'version' above all others", func() {
			result, ok := versions.Extract("abc-1.0.5-def-2.3.4.tgz", "abc-(.*)-def-(?P<version>.*).tgz")
			Expect(ok).To(BeTrue())

			Expect(result.Path).To(Equal("abc-1.0.5-def-2.3.4.tgz"))
			Expect(result.Version.String()).To(Equal("2.3.4"))
			Expect(result.VersionNumber).To(Equal("2.3.4"))
		})
	})
})
