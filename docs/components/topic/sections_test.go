package topic

import (
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		intro    string
		sections []Section
	}{
		{
			name:   "intro only",
			source: "Some text.\n\n### Deeper\n",
			intro:  "Some text.\n\n### Deeper\n",
		},
		{
			name:   "two sections",
			source: "Lead.\n\n## Values\n\nFirst.\n## Rules\nSecond.\n",
			intro:  "Lead.\n\n",
			sections: []Section{
				{Slug: "values", Title: "Values", Source: "## Values\n\nFirst.\n"},
				{Slug: "rules", Title: "Rules", Source: "## Rules\nSecond.\n"},
			},
		},
		{
			name:   "a heading inside a backtick fence is not a split",
			source: "## Code\n```md\n## Not a section\n```\nAfter.\n",
			sections: []Section{
				{Slug: "code", Title: "Code", Source: "## Code\n```md\n## Not a section\n```\nAfter.\n"},
			},
		},
		{
			name:   "a heading inside a tilde fence is not a split",
			source: "## Code\n~~~\n## Not a section\n```\n## Still code\n~~~\n## Next\n",
			sections: []Section{
				{Slug: "code", Title: "Code", Source: "## Code\n~~~\n## Not a section\n```\n## Still code\n~~~\n"},
				{Slug: "next", Title: "Next", Source: "## Next\n"},
			},
		},
		{
			name:   "deeper headings stay inside the section",
			source: "## Rule\n### Building\nText.\n#### Detail\n",
			sections: []Section{
				{Slug: "rule", Title: "Rule", Source: "## Rule\n### Building\nText.\n#### Detail\n"},
			},
		},
		{
			name:   "duplicate titles get numbered slugs",
			source: "## CSS\na\n## CSS\nb\n## CSS\nc\n",
			sections: []Section{
				{Slug: "css", Title: "CSS", Source: "## CSS\na\n"},
				{Slug: "css-2", Title: "CSS", Source: "## CSS\nb\n"},
				{Slug: "css-3", Title: "CSS", Source: "## CSS\nc\n"},
			},
		},
		{
			name:   "punctuation collapses into dashes and is trimmed",
			source: "## CSS()\n## Layout and Screen interfaces\n## Example: fetching & displaying!\n## ???\n",
			sections: []Section{
				{Slug: "css", Title: "CSS()", Source: "## CSS()\n"},
				{Slug: "layout-and-screen-interfaces", Title: "Layout and Screen interfaces", Source: "## Layout and Screen interfaces\n"},
				{Slug: "example-fetching-displaying", Title: "Example: fetching & displaying!", Source: "## Example: fetching & displaying!\n"},
				{Slug: "section", Title: "???", Source: "## ???\n"},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			intro, sections := Split(test.source)
			if intro != test.intro {
				t.Errorf("intro = %q, want %q", intro, test.intro)
			}
			if !slices.Equal(sections, test.sections) {
				t.Errorf("sections = %#v, want %#v", sections, test.sections)
			}
		})
	}
}
