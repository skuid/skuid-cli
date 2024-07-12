package testutil

import (
	"testing"

	utilmocks "github.com/skuid/skuid-cli/pkg/util/mocks"
	"github.com/stretchr/testify/mock"
)

/*
NOTE - Mock Builders contain "Without*" methods because "github.com/stretchr/testify/mock" does not contain a reliable
way to clear/reset expectations.  It does offer the "Unset" API for a "Call", however it has several issues and could
result in unexpected behavior if used.  In order to support a "default" Mock via the Builder but override certain
expectations, the Builders "Without*" methods should be called.  This will cause the builder to not set any expectations
for the specified method so that they can then be setup within a test.  If/When Issue #1621 (see link below) is addressed,
the "Without*" methods can be eliminated and then each test can unset/clear/reset individual expectations.

See the following issues related to Unset
- https://github.com/stretchr/testify/issues/1621
- https://github.com/stretchr/testify/issues/1599
- https://github.com/stretchr/testify/issues/1542

See the following PRs for activity on Unset
- https://github.com/stretchr/testify/pull/1622
- https://github.com/stretchr/testify/pull/1605
- https://github.com/stretchr/testify/pull/1572
- https://github.com/stretchr/testify/pull/1511
- https://github.com/stretchr/testify/pull/1425
- https://github.com/stretchr/testify/pull/1239
*/

type ZipFileBuilder struct {
	skipWrite bool
}

func (b *ZipFileBuilder) WithoutWrite() *ZipFileBuilder {
	b.skipWrite = true
	return b
}

func (b *ZipFileBuilder) Build(t *testing.T) *utilmocks.ZipFile {
	mockZipFile := utilmocks.NewZipFile(t)

	if !b.skipWrite {
		mockZipFile.EXPECT().Write(mock.Anything).RunAndReturn(func(b []byte) (int, error) {
			return len(b), nil
		})
	}

	return mockZipFile
}

func NewZipFileBuilder() *ZipFileBuilder {
	return &ZipFileBuilder{}
}
