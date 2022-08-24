package cmd

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/efficientgo/tools/core/pkg/testutil"
)

func TestHandleGraphs(t *testing.T) {
	tc := []string{"go_gc_duration_seconds", "prometheus_engine_query_duration_seconds"}
	wd, err := os.Getwd()
	testutil.Ok(t, err)
	dir := t.TempDir()

	t.Run("ascii graph", func(t *testing.T) {
		for _, n := range tc {
			out := bytes.NewBufferString("")

			ib, err := os.ReadFile(wd + "/testdata/" + n + ".json")
			testutil.Ok(t, err)

			testutil.Ok(t, handleGraph(ib, "ascii", n, path.Join(dir, "test.png"), out))

			exp, err := os.ReadFile(wd + "/testdata/" + n + ".txt")
			testutil.Ok(t, err)

			testutil.Equals(t, exp, out.Bytes())
		}
	})

	t.Run("png graph", func(t *testing.T) {
		for _, n := range tc {
			ib, err := os.ReadFile(wd + "/testdata/" + n + ".json")
			testutil.Ok(t, err)

			testutil.Ok(t, handleGraph(ib, "png", n, path.Join(dir, "test.png"), bytes.NewBufferString("")))

			exp, err := os.ReadFile(wd + "/testdata/" + n + ".png")
			testutil.Ok(t, err)

			out, err := os.ReadFile(path.Join(dir, "test.png"))
			testutil.Ok(t, err)

			testutil.Equals(t, exp, out)
		}
	})
}
