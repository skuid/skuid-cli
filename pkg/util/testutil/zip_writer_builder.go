package testutil

import (
	"strings"
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

type ZipWriterBuilder struct {
	skipCreate bool
	skipClose  bool
}

func (b *ZipWriterBuilder) WithoutCreate() *ZipWriterBuilder {
	b.skipCreate = true
	return b
}

func (b *ZipWriterBuilder) WithoutClose() *ZipWriterBuilder {
	b.skipClose = true
	return b
}

func (b *ZipWriterBuilder) Build(t *testing.T) *utilmocks.ZipWriter {
	mockZipWriter := utilmocks.NewZipWriter(t)

	if !b.skipCreate {
		mockZipWriter.EXPECT().Create(mock.Anything).Return(&strings.Builder{}, nil).Maybe()
	}

	if !b.skipClose {
		mockZipWriter.EXPECT().Close().Return(nil).Maybe()
	}

	return mockZipWriter
}

func NewZipWriterBuilder() *ZipWriterBuilder {
	return &ZipWriterBuilder{}
}
