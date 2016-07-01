/***** BEGIN LICENSE BLOCK *****
#
# ***** END LICENSE BLOCK *****/

package arraysplitter

import (
	"bytes"

	"github.com/mozilla-services/heka/pipeline"
	gs "github.com/rafrombrc/gospec/src/gospec"
)

func makeSplitterRunner(name string, splitter pipeline.Splitter) pipeline.SplitterRunner {
	srConfig := pipeline.CommonSplitterConfig{}
	return pipeline.NewSplitterRunner(name, splitter, srConfig)
}

func ArraySpec(c gs.Context) {
	c.Specify("A ArraySplitter", func() {
		splitter := &ArraySplitter{}
		config := splitter.ConfigStruct().(*ArraySplitterConfig)
		config.Delimiter = []byte{'\n'}
		config.CheckSumByteCount = 0
		config.IndexByteCount = 0
		config.IndexByteOrder = "big"
		config.LengthByteCount = 0
		config.LengthByteOrder = "b"
		sRunner := makeSplitterRunner("ArraySplitter", splitter)
		buf := []byte("test1\ntest12\ntest123\npartial")

		c.Specify("using default delimiter", func() {
			err := splitter.Init(config)
			c.Assume(err, gs.IsNil)
			reader := bytes.NewReader(buf)
			n, record, err := sRunner.GetRecordFromStream(reader)
			c.Expect(n, gs.Equals, 6)
			c.Expect(err, gs.IsNil)
			c.Expect(string(record), gs.Equals, "test1")
			n, record, err = sRunner.GetRecordFromStream(reader)
			c.Expect(n, gs.Equals, 7)
			c.Expect(err, gs.IsNil)
			c.Expect(string(record), gs.Equals, "test12")
			n, record, err = sRunner.GetRecordFromStream(reader)
			c.Expect(n, gs.Equals, 8)
			c.Expect(err, gs.IsNil)
			c.Expect(string(record), gs.Equals, "test123")
			n, record, err = sRunner.GetRecordFromStream(reader)
			c.Expect(n, gs.Equals, 0)
			c.Expect(err, gs.IsNil)
			c.Expect(string(sRunner.GetRemainingData()), gs.Equals, "partial")
		})

		c.Specify("using tab delimiter", func() {
			reader := bytes.NewReader([]byte("test1\ttest2\t"))
			config.Delimiter = []byte{'\t'}
			err := splitter.Init(config)
			c.Assume(err, gs.IsNil)
			n, record, err := sRunner.GetRecordFromStream(reader)
			c.Expect(n, gs.Equals, 6)
			c.Expect(err, gs.IsNil)
			c.Expect(string(record), gs.Equals, "test1")
			n, record, err = sRunner.GetRecordFromStream(reader)
			c.Expect(n, gs.Equals, 6)
			c.Expect(err, gs.IsNil)
			c.Expect(string(record), gs.Equals, "test2")
		})
	})
}
