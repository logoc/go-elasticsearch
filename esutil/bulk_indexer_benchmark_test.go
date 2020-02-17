// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

// +build !integration

package esutil_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type mockTransp struct{}

func (t *mockTransp) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
}

func BenchmarkBulkIndexer(b *testing.B) {
	b.ReportAllocs()

	b.Run("Basic", func(b *testing.B) {
		b.ResetTimer()

		es, _ := elasticsearch.NewClient(elasticsearch.Config{Transport: &mockTransp{}})
		bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Client:     es,
			FlushBytes: 1024,
		})
		defer bi.Close(context.Background())

		docID := make([]byte, 0, 16)
		var docIDBuf bytes.Buffer
		docIDBuf.Grow(cap(docID))

		for i := 0; i < b.N; i++ {
			docID = strconv.AppendInt(docID, int64(i), 10)
			docIDBuf.Write(docID)
			bi.Add(context.Background(), esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: docIDBuf.String(),       // 1x alloc
				Body:       strings.NewReader(`{}`), // 1x alloc
			})
			docID = docID[:0]
			docIDBuf.Reset()
		}
	})
}
