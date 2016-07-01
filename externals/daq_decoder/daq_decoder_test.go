/***** BEGIN LICENSE BLOCK *****
#
# ***** END LICENSE BLOCK *****/

package daqdecoder

import (
	. "github.com/mozilla-services/heka/pipeline"
	pipeline_ts "github.com/mozilla-services/heka/pipeline/testsupport"
	"github.com/mozilla-services/heka/pipelinemock"
	"github.com/rafrombrc/gomock/gomock"
	"github.com/rafrombrc/gospec/src/gospec"
	gs "github.com/rafrombrc/gospec/src/gospec"
)

func DaqDecoderSpec(c gospec.Context) {
	t := &pipeline_ts.SimpleT{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	globals := &GlobalConfigStruct{
		PluginChanSize: 5,
	}
	config := NewPipelineConfig(globals)

	c.Specify("A DaqDecoder", func() {
		decoder := new(DaqDecoder)

		conf := decoder.ConfigStruct().(*DaqDecoderConfig)

		conf.Frames = map[string]frameConfig{
			"power": frameConfig{2,
				[][]string{
					[]string{"pv", "u2"},
					[]string{"pi", "u2"},
					[]string{"nv", "u2"},
					[]string{"ni", "u2"},
				}},
			"fog": frameConfig{1,
				[][]string{
					[]string{"idx", "u1"},
					[]string{"data", "i4"},
					[]string{"t1", "u1"},
					[]string{"t2", "u1"},
					[]string{"st", "u1"},
				}},
		}
		err := decoder.Init(conf)
		c.Assume(err, gs.IsNil)

		dRunner := pipelinemock.NewMockDecoderRunner(ctrl)
		decoder.SetDecoderRunner(dRunner)
		pack := NewPipelinePack(config.InputRecycleChan())

		c.Specify("correctly converts FogData to fields", func() {
			pl := []byte{1, 1, 0, 9, 1, 1, 2, 3, 4}
			pack.Message.SetPayload(string(pl))
			_, err := decoder.Decode(pack)
			c.Expect(err, gs.IsNil)

			value, ok := pack.Message.GetFieldValue("idx")
			c.Expect(ok, gs.IsTrue)
			expected := int64(1)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("data")
			c.Expect(ok, gs.IsTrue)
			expected = int64(0*256*256*256 + 9*256*256 + 1*256 + 1)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("t1")
			c.Expect(ok, gs.IsTrue)
			expected = int64(2)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("t2")
			c.Expect(ok, gs.IsTrue)
			expected = int64(3)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("st")
			c.Expect(ok, gs.IsTrue)
			expected = int64(4)
			c.Expect(value.(int64), gs.Equals, expected)

			value = pack.Message.GetType()
			c.Expect(value, gs.Equals, "fog")
		})

		c.Specify("correctly converts PowerData to fields", func() {
			pl := []byte{2, 0, 6, 0, 2, 0, 5, 0, 1}
			pack.Message.SetPayload(string(pl))
			_, err := decoder.Decode(pack)
			c.Expect(err, gs.IsNil)

			value, ok := pack.Message.GetFieldValue("pv")
			c.Expect(ok, gs.IsTrue)
			expected := int64(6)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("pi")
			c.Expect(ok, gs.IsTrue)
			expected = int64(2)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("nv")
			c.Expect(ok, gs.IsTrue)
			expected = int64(5)
			c.Expect(value.(int64), gs.Equals, expected)

			value, ok = pack.Message.GetFieldValue("ni")
			c.Expect(ok, gs.IsTrue)
			expected = int64(1)
			c.Expect(value.(int64), gs.Equals, expected)

			value = pack.Message.GetType()
			c.Expect(value, gs.Equals, "power")
		})
	})
}
