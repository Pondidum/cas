package command

import (
	"io"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
)

func TestHashingSingleFileFromStdin(t *testing.T) {

	cases := []struct {
		name         string
		files        map[string]string
		stdin        string
		arg          string
		expectedHash string
	}{
		{
			name: "stdin: hashing a single file",
			files: map[string]string{
				"main.go": "some content",
			},
			stdin:        "main.go\n",
			expectedHash: "ffa798d14d7ac63881d209f113750bbeac9f2c652582f9681e8a59324c204ea0",
		},
		{
			name: "stdin: hashing the same file with different content",
			files: map[string]string{
				"main.go": "some different content",
			},
			stdin:        "main.go\n",
			expectedHash: "be5dbb2fa4fbe52b98dcc976592b6d1276038638d23bbba61cf11597055780bc",
		},
		{
			name: "stdin: hashing a different file with the same content",
			files: map[string]string{
				"different.go": "some different content",
			},
			stdin:        "different.go\n",
			expectedHash: "e1cc8748f0c92db0ca91dfa146381084122d63fa8f085f1685ae75b16e860c1a",
		},
		{
			name: "stdin: hashing multiple files",
			files: map[string]string{
				"main.go":      "some content",
				"different.go": "some different content",
			},
			stdin:        "main.go\ndifferent.go\n",
			expectedHash: "0a7b98045af4d434e5a84f04d5391aa6ef265def4e35d0cb47d3da7270d6943e",
		},
		{
			name: "stdin: hashing multiple files in a different order",
			files: map[string]string{
				"main.go":      "some content",
				"different.go": "some different content",
			},
			stdin:        "different.go\nmain.go\n",
			expectedHash: "0a7b98045af4d434e5a84f04d5391aa6ef265def4e35d0cb47d3da7270d6943e",
		},
		{
			name: "stdin: hashing multiple files with different content",
			files: map[string]string{
				"main.go":      "some other content",
				"different.go": "some different content",
			},
			stdin:        "main.go\ndifferent.go\n",
			expectedHash: "e9ddf7e9893c19b4f2b924e7774036847b69bb6f7d372810917aa755d2cbf3af",
		},
		{
			name: "file arg: hashing a single file",
			files: map[string]string{
				"main.go":  "some content",
				"filelist": "main.go\n",
			},
			arg:          "filelist",
			expectedHash: "ffa798d14d7ac63881d209f113750bbeac9f2c652582f9681e8a59324c204ea0",
		},
		{
			name: "file arg: hashing the same file with different content",
			files: map[string]string{
				"main.go":  "some different content",
				"filelist": "main.go\n",
			},
			arg:          "filelist",
			expectedHash: "be5dbb2fa4fbe52b98dcc976592b6d1276038638d23bbba61cf11597055780bc",
		},
		{
			name: "file arg: hashing a different file with the same content",
			files: map[string]string{
				"different.go": "some different content",
				"filelist":     "different.go\n",
			},
			arg:          "filelist",
			expectedHash: "e1cc8748f0c92db0ca91dfa146381084122d63fa8f085f1685ae75b16e860c1a",
		},
		{
			name: "file arg: hashing multiple files",
			files: map[string]string{
				"main.go":      "some content",
				"different.go": "some different content",
				"filelist":     "main.go\ndifferent.go\n",
			},
			arg:          "filelist",
			expectedHash: "0a7b98045af4d434e5a84f04d5391aa6ef265def4e35d0cb47d3da7270d6943e",
		},
		{
			name: "file arg: hashing multiple files in a different order",
			files: map[string]string{
				"main.go":      "some content",
				"different.go": "some different content",
				"filelist":     "different.go\nmain.go\n",
			},
			arg:          "filelist",
			expectedHash: "0a7b98045af4d434e5a84f04d5391aa6ef265def4e35d0cb47d3da7270d6943e",
		},
		{
			name: "file arg: hashing multiple files with different content",
			files: map[string]string{
				"main.go":      "some other content",
				"different.go": "some different content",
				"filelist":     "main.go\ndifferent.go\n",
			},
			arg:          "filelist",
			expectedHash: "e9ddf7e9893c19b4f2b924e7774036847b69bb6f7d372810917aa755d2cbf3af",
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			fs := memfs.New()
			for n, c := range tc.files {
				fs.WriteFile(n, []byte(c), 0755)
			}

			ui := cli.NewMockUi()
			cmd := &HashCommand{fs: fs}
			cmd.Meta = NewMeta(ui, cmd)

			if tc.stdin != "" {
				cmd.testInput = io.NopCloser(strings.NewReader(tc.stdin))
			}

			assert.Equal(t, 0, cmd.Run([]string{tc.arg}), ui.ErrorWriter.String())
			assert.Equal(t, tc.expectedHash+"\n", ui.OutputWriter.String())
		})
	}

}
