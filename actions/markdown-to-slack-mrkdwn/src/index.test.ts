import { describe, expect, it } from "bun:test";
import { transform } from "./index";

// Corpus pulled from upstream slackify-markdown
// (https://github.com/jsarafajr/slackify-markdown/blob/master/__test__/slackify-markdown.test.ts)
// so this action is locked to the same behavior as the library it wraps.
const zws = String.fromCharCode(0x200b);

const cases: [name: string, input: string, expected: string][] = [
  ["Simple text", "hello world", "hello world\n"],
  ["Escaped text", "*h&ello>world<", "*h&amp;ello&gt;world&lt;\n"],
  [
    "Definitions",
    "hello\n\n[1]: http://atlassian.com\n\nworld\n\n[2]: http://atlassian.com",
    "hello\n\nworld\n",
  ],
  [
    "Headings",
    "# heading 1\n## heading 2\n### heading 3",
    "*heading 1*\n\n*heading 2*\n\n*heading 3*\n",
  ],
  ["Bold", "**bold text**", `${zws}*bold text*${zws}\n`],
  ["Bold character in word", "he**l**lo", `he${zws}*l*${zws}lo\n`],
  ["Italic", "*italic text*", `${zws}_italic text_${zws}\n`],
  [
    "Bold+Italic",
    "***bold+italic***",
    `${zws}_${zws}*bold+italic*${zws}_${zws}\n`,
  ],
  ["Strike", "~~strike text~~", `${zws}~strike text~${zws}\n`],
  [
    "Unordered list",
    "* list\n* list\n* list",
    "•   list\n•   list\n•   list\n",
  ],
  [
    "Ordered list",
    "1. list\n2. list\n3. list",
    "1.  list\n2.  list\n3.  list\n",
  ],
  [
    "Link with title",
    '[](http://atlassian.com "Atlassian")',
    "<http://atlassian.com|Atlassian>\n",
  ],
  [
    "Link with alt",
    "[test](http://atlassian.com)",
    "<http://atlassian.com|test>\n",
  ],
  [
    "Link with alt and title",
    '[test](http://atlassian.com "Atlassian")',
    "<http://atlassian.com|test>\n",
  ],
  [
    "Link with angle bracket syntax",
    "<http://atlassian.com>",
    "<http://atlassian.com|http://atlassian.com>\n",
  ],
  [
    "Link with no alt nor title",
    "[](http://atlassian.com)",
    "<http://atlassian.com>\n",
  ],
  ["Link with invalid URL", "[test](/atlassian)", "test\n"],
  [
    "Link in reference style with alt",
    "[Atlassian]\n\n[atlassian]: http://atlassian.com",
    "<http://atlassian.com|Atlassian>\n",
  ],
  [
    "Link in reference style with custom label",
    "[][test]\n\n[test]: http://atlassian.com",
    "<http://atlassian.com>\n",
  ],
  [
    "Link in reference style with alt and custom label",
    "[Atlassian][test]\n\n[test]: http://atlassian.com",
    "<http://atlassian.com|Atlassian>\n",
  ],
  [
    "Link in reference style with title",
    '[][test]\n\n[test]: http://atlassian.com "Title"',
    "<http://atlassian.com|Title>\n",
  ],
  [
    "Link in reference style with alt and title",
    '[Atlassian]\n\n[atlassian]: http://atlassian.com "Title"',
    "<http://atlassian.com|Atlassian>\n",
  ],
  [
    "Link is already encoded",
    "[Atlassian](https://www.atlassian.com?redirect=https%3A%2F%2Fwww.asana.com): /atlassian",
    "<https://www.atlassian.com?redirect=https%3A%2F%2Fwww.asana.com|Atlassian>: /atlassian\n",
  ],
  [
    "Link in reference style with invalid definition",
    "[Atlassian][test]\n\n[test]: /atlassian",
    "Atlassian\n",
  ],
  [
    "Image with title",
    '![](https://bitbucket.org/repo/123/images/logo.png "test")',
    "<https://bitbucket.org/repo/123/images/logo.png|test>\n",
  ],
  [
    "Image with alt",
    "![logo.png](https://bitbucket.org/repo/123/images/logo.png)",
    "<https://bitbucket.org/repo/123/images/logo.png|logo.png>\n",
  ],
  [
    "Image with alt and title",
    "![logo.png](https://bitbucket.org/repo/123/images/logo.png 'test')",
    "<https://bitbucket.org/repo/123/images/logo.png|logo.png>\n",
  ],
  [
    "Image with no alt nor title",
    "![](https://bitbucket.org/repo/123/images/logo.png)",
    "<https://bitbucket.org/repo/123/images/logo.png>\n",
  ],
  [
    "Image with invalid URL",
    "![logo.png](/relative-path-logo.png 'test')",
    "logo.png\n",
  ],
  [
    "Image in reference style with alt",
    "![Atlassian]\n\n[atlassian]: https://bitbucket.org/repo/123/images/logo.png",
    "<https://bitbucket.org/repo/123/images/logo.png|Atlassian>\n",
  ],
  [
    "Image in reference style with custom label",
    "![][test]\n\n[test]: https://bitbucket.org/repo/123/images/logo.png",
    "<https://bitbucket.org/repo/123/images/logo.png>\n",
  ],
  [
    "Image in reference style with alt and custom label",
    "![Atlassian][test]\n\n[test]: https://bitbucket.org/repo/123/images/logo.png",
    "<https://bitbucket.org/repo/123/images/logo.png|Atlassian>\n",
  ],
  [
    "Image in reference style with title",
    '![][test]\n\n[test]: https://bitbucket.org/repo/123/images/logo.png "Title"',
    "<https://bitbucket.org/repo/123/images/logo.png|Title>\n",
  ],
  [
    "Image in reference style with alt and title",
    '![Atlassian]\n\n[atlassian]: https://bitbucket.org/repo/123/images/logo.png "Title"',
    "<https://bitbucket.org/repo/123/images/logo.png|Atlassian>\n",
  ],
  [
    "Image in reference style with invalid definition",
    "![Atlassian][test]\n\n[test]: /relative-path-logo.png",
    "Atlassian\n",
  ],
  ["Inline code", "hello `world`", "hello `world`\n"],
  ["Code block", "```\ncode block\n```", "```\ncode block\n```\n"],
  [
    "Code block with newlines",
    "```\ncode\n\n\nblock\n```",
    "```\ncode\n\n\nblock\n```\n",
  ],
  [
    "Code block with language",
    "```javascript\ncode block\n```",
    "```\ncode block\n```\n",
  ],
  [
    "Code block with deprecated language declaration",
    "```\n#!javascript\ncode block\n```",
    "```\ncode block\n```\n",
  ],
  ["User mention", "<@UPXGB22A2>", "<@UPXGB22A2>\n"],
  ["Channel mention", "<#C04A9JK5R3Z>", "<#C04A9JK5R3Z>\n"],
  [
    "Blockquote - single line",
    "> This is a blockquote",
    "> This is a blockquote\n",
  ],
  [
    "Blockquote - multi-line",
    "> This is the first line\nThis is the second line\nThis is the third line",
    "> This is the first line\n> This is the second line\n> This is the third line\n",
  ],
  [
    "Blockquote - with inline code",
    "> This has `inline code` inside",
    "> This has `inline code` inside\n",
  ],
  [
    "Blockquote - with multiple paragraphs",
    "> First paragraph\n>\n> Second paragraph",
    "> First paragraph\n\n> Second paragraph\n",
  ],
  [
    "Blockquote - with formatted text",
    "> This has **bold**, *italic*, and a [link](http://example.com)",
    `> This has ${zws}*bold*${zws}, ${zws}_italic_${zws}, and a <http://example.com|link>\n`,
  ],
  [
    "HTML comment - single line",
    "<!-- comment text -->\n\n## Heading\nContent",
    "*Heading*\n\nContent\n",
  ],
  [
    "HTML comment - multi-line",
    "<!--\nRelease notes\ngenerated automatically\n-->\n\n## Foo\nbar",
    "*Foo*\n\nbar\n",
  ],
];

describe("transform", () => {
  for (const [name, input, expected] of cases) {
    it(name, () => {
      expect(transform(input)).toBe(expected);
    });
  }

  it("handles null/undefined", () => {
    expect(transform(null)).toBe("");
    expect(transform(undefined)).toBe("");
  });
});
