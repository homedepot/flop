package flop

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsDirChecksDirCorrectlyWhenIsDirFieldSet(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing_file", tmpFile(), false},
		{"existing_dir", tmpDirPath(), true},
		{"non_existing_path", tmpFilePathUnused(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := File{Path: tt.path}
			assert.Equal(tt.expected, f.IsDir())
		})
	}
}
