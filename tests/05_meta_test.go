package tests

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ncw/swift/v2"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetaContainer(t *testing.T) {
	c, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	err := c.Authenticate(ctx)
	assert.NoError(t, err)

	//

	err = c.ContainerCreate(ctx, "Xcontainer", swift.Headers{})
	assert.NoError(t, err)

	//

	m := swift.Metadata{"color": "orange"}
	headers := m.ContainerHeaders()
	err = c.ContainerUpdate(ctx, "Xcontainer", headers)
	assert.NoError(t, err)

	//

	_, headers, err = c.Container(ctx, "Xcontainer")
	assert.NoError(t, err)
	assert.Equal(t, m, headers.ContainerMetadata())

}

func TestUpdateMetaObject(t *testing.T) {
	c, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	err := c.Authenticate(ctx)
	assert.NoError(t, err)

	//

	err = c.ContainerCreate(ctx, "Xcontainer", swift.Headers{})
	assert.NoError(t, err)

	content := time.Now().Format(time.RFC3339)
	err = c.ObjectPutString(ctx, "Xcontainer", "a1/b2/c3.txt", content, "text/plain")
	assert.NoError(t, err)

	//

	info, headers, err := c.Object(ctx, "Xcontainer", "a1/b2/c3.txt")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), info.Bytes)
	assert.Equal(t, "text/plain", info.ContentType)
	assert.Equal(t, fmt.Sprintf("%x", md5.Sum([]byte(content))), info.Hash)
	assert.Equal(t, swift.Metadata{}, headers.ContainerMetadata())

	//

	m := swift.Metadata{"color": "orange"}
	headers = m.ObjectHeaders()
	err = c.ObjectUpdate(ctx, "Xcontainer", "a1/b2/c3.txt", headers)
	assert.NoError(t, err)

	//

	info, headers, err = c.Object(ctx, "Xcontainer", "a1/b2/c3.txt")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), info.Bytes)
	assert.Equal(t, "text/plain", info.ContentType)
	assert.Equal(t, fmt.Sprintf("%x", md5.Sum([]byte(content))), info.Hash)
	assert.Equal(t, m, headers.ObjectMetadata())
}

func TestUpdateMetaLargeObject(t *testing.T) {
	c, cleanup := setup()
	defer cleanup()

	ctx := context.Background()
	err := c.Authenticate(ctx)
	assert.NoError(t, err)

	//

	err = c.ContainerCreate(ctx, "Xcontainer", swift.Headers{})
	assert.NoError(t, err)
	err = c.ContainerCreate(ctx, "Chunks-Container", swift.Headers{})
	assert.NoError(t, err)

	content := strings.Repeat(time.Now().Format(time.RFC3339), 10<<20) // 250MiB

	dlo, err := c.DynamicLargeObjectCreate(ctx, &swift.LargeObjectOpts{
		Container:          "Xcontainer",
		ObjectName:         "a42/dates.txt",
		CheckHash:          true, // Hash of each chunk object
		SegmentContainer:   "Chunks-Container",
		SegmentPrefix:      "a42",
		ChunkSize:          100 << 20, // 100 MiB
	})
	assert.NoError(t, err)

	n, err := io.Copy(dlo, bytes.NewBufferString(content))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), n)

	err = dlo.Flush(ctx)
	assert.NoError(t, err)
	err = dlo.Close()
	assert.NoError(t, err)

	//

	info, headers, err := c.Object(ctx, "Xcontainer", "a42/dates.txt")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), info.Bytes)
	assert.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	assert.NotEmpty(t, info.Hash)
	assert.Equal(t, swift.Metadata{}, headers.ContainerMetadata())

	//

	m := swift.Metadata{"color": "orange"}
	headers = m.ObjectHeaders()
	headers["Content-Type"] = "text/plain; charset=utf-8"
	err = c.ObjectUpdate(ctx, "Xcontainer", "a42/dates.txt", headers)
	assert.NoError(t, err)

	//

	info, headers, err = c.Object(ctx, "Xcontainer", "a42/dates.txt")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), info.Bytes)
	assert.Equal(t, "text/plain; charset=utf-8", info.ContentType)
	assert.NotEmpty(t, info.Hash)
	assert.Equal(t, m, headers.ObjectMetadata())
}
