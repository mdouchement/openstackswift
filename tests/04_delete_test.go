package tests

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ncw/swift/v2"
	"github.com/stretchr/testify/assert"
)

func TestDeleteObject(t *testing.T) {
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

	err = c.ObjectDelete(ctx, "Xcontainer", "a1/b2/c3.txt")
	assert.NoError(t, err)

	//

	_, _, err = c.Object(ctx, "Xcontainer", "a1/b2/c3.txt")
	assert.Error(t, swift.ObjectNotFound, err)
}

func TestDeleteLargeObject(t *testing.T) {
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
		Container:        "Xcontainer",
		ObjectName:       "a42/dates.txt",
		CheckHash:        true, // Hash of each chunk object
		SegmentContainer: "Chunks-Container",
		SegmentPrefix:    "a42",
		ChunkSize:        100 << 20, // 100 MiB
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

	err = c.ObjectDelete(ctx, "Xcontainer", "a42/dates.txt")
	assert.NoError(t, err)

	//

	_, _, err = c.Object(ctx, "Xcontainer", "a42/dates.txt")
	assert.Error(t, swift.ObjectNotFound, err)

	//

	objects, err := c.Objects(ctx, "Chunks-Container", &swift.ObjectsOpts{})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(objects))
}
