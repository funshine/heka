/***** BEGIN LICENSE BLOCK *****
#
# ***** END LICENSE BLOCK *****/

package daqdecoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"reflect"
	"strconv"
	"strings"
)

const (
	portConfig        = 0x00
	portFogData       = 0x01
	portPowerData     = 0x02
	portPowerOn       = 0x03
	portPowerOff      = 0x04
	portTempData      = 0x05
	portFogandpwrData = 0x06
	slotNum           = 9
)

type frameConfig struct {
	FrameType   int        `toml:"frame_type"`
	FramePatern [][]string `toml:"frame_patern"`
}

type DaqDecoderConfig struct {
	// Regular expression that describes log line format and capture group
	// values.
	Frames map[string]frameConfig
}

type item struct {
	name string
	v    interface{}
}

type frame struct {
	fameName    string
	framePatern []item
}

type frameType int

type DaqDecoder struct {
	runner DecoderRunner
	frames map[frameType]frame
}

// decodePatern takes a string slice such like: ["index", "i1"], ["data", "i4"], ["", "n4"].
// and returns type item with corresponding name type(byte length emebedded), and error if error occurs.
// ["index", "i1"] -> &item{"index", int8(0)}, nil
// ["data", "i4"] -> &item{"data", int32(0)}, nil
// ["", "n4"] -> &item{"", []byte{0,0,0,0} }, nil
// ["data1", "n-5"] -> nil, error
func (d *DaqDecoder) decodePatern(s []string) (it *item, err error) {
	if len(s) < 2 {
		return nil, fmt.Errorf("Frame item %s should has 2 strings", s)
	}
	if s[1] == "" {
		return nil, fmt.Errorf("Frame item %s should has format string", s)
	}
	sl := strings.ToLower(s[1])
	switch sl {
	case "i1", "int8":
		return &item{s[0], int8(0)}, nil
	case "i2", "int16":
		return &item{s[0], int16(0)}, nil
	case "i4", "int32":
		return &item{s[0], int32(0)}, nil
	case "i8", "int64":
		return &item{s[0], int64(0)}, nil
	case "u1", "uint8":
		return &item{s[0], uint8(0)}, nil
	case "u2", "uint16":
		return &item{s[0], uint16(0)}, nil
	case "u4", "uint32":
		return &item{s[0], uint32(0)}, nil
	case "u8", "uint64":
		return &item{s[0], uint64(0)}, nil
	case "f4", "float32":
		return &item{s[0], float32(0)}, nil
	case "f8", "float64":
		return &item{s[0], float64(0)}, nil
	default:
		sn := sl
		nm := s[0]
		if strings.HasPrefix(sl, "n") {
			sn = strings.TrimPrefix(sl, "n")
			nm = ""
		} else if strings.HasPrefix(sl, "omit") {
			sn = strings.TrimPrefix(sl, "omit")
			nm = ""
		} else if strings.HasPrefix(sl, "byte") {
			sn = strings.TrimPrefix(sl, "byte")
		} else if strings.HasPrefix(sl, "b") {
			sn = strings.TrimPrefix(sl, "b")
		} else {
			return nil, fmt.Errorf("Format %s not supported", s[1])
		}
		num, err := strconv.Atoi(sn)
		if num < 0 {
			return nil, fmt.Errorf("%s: Omitted number should be >= 0", s[1])
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %s", s, err)
		}
		return &item{nm, make([]byte, num)}, nil
	}
}

func (d *DaqDecoder) ConfigStruct() interface{} {
	return &DaqDecoderConfig{}
}

func (d *DaqDecoder) Init(config interface{}) error {
	conf := config.(*DaqDecoderConfig)
	d.frames = make(map[frameType]frame)
	for name, f := range conf.Frames {
		framePatern := make([]item, 0)
		for _, v := range f.FramePatern {
			vit, err := d.decodePatern(v)
			if err != nil {
				return err
			}
			framePatern = append(framePatern, *vit)
		}
		fr := frame{name, framePatern}
		d.frames[frameType(f.FrameType)] = fr
	}

	for k, v := range d.frames {
		fmt.Println(v.fameName, ": type:", k)
		for kk, vv := range v.framePatern {
			fmt.Println(kk, vv.name, reflect.TypeOf(vv.v))
		}
	}
	return nil
}

// Implement `WantsDecoderRunner`
func (d *DaqDecoder) SetDecoderRunner(dr DecoderRunner) {
	d.runner = dr
}

func (d *DaqDecoder) Decode(pack *PipelinePack) (packs []*PipelinePack,
	err error) {
	pl := []byte(pack.Message.GetPayload())

	if len(pl) == 0 {
		err = fmt.Errorf("nothing in payload to decode")
		return
	}
	buf := bytes.NewBuffer(pl)

	var dataType uint8
	if err = binary.Read(buf, binary.BigEndian, &dataType); err != nil {
		return
	}

	fr, exists := d.frames[frameType(dataType&0x0F)]
	if !exists {
		err = fmt.Errorf("packet type not supported")
		return
	}

	for _, it := range fr.framePatern {
		switch t := it.v.(type) {
		case int8:
			v := it.v.(int8)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case int16:
			v := it.v.(int16)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case int32:
			v := it.v.(int32)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case int64:
			v := it.v.(int64)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case uint8:
			v := it.v.(uint8)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case uint16:
			v := it.v.(uint16)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case uint32:
			v := it.v.(uint32)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case uint64:
			v := it.v.(uint64)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, int64(v))
			}
		case float32:
			v := it.v.(float32)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, float64(v))
			}
		case float64:
			v := it.v.(float64)
			if err = binary.Read(buf, binary.BigEndian, &v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, float64(v))
			}
		case []byte:
			v := it.v.([]byte)
			if err = binary.Read(buf, binary.BigEndian, v); err != nil {
				return
			}
			if it.name != "" {
				err = d.addStatField(pack, it.name, []byte(v))
			}
		default:
			err = fmt.Errorf("Format %s not supported", t)
		}
		if err != nil {
			return
		}
	}

	pack.Message.SetType(fr.fameName)
	pack.Message.SetPayload("")

	packs = []*PipelinePack{pack}
	return
}

func (d *DaqDecoder) addStatField(pack *PipelinePack, name string,
	value interface{}) error {

	field, err := message.NewField(name, value, "")
	if err != nil {
		return fmt.Errorf("error adding field '%s': %s", name, err)
	}
	pack.Message.AddField(field)
	return nil
}

func init() {
	RegisterPlugin("DaqDecoder", func() interface{} {
		return new(DaqDecoder)
	})
}
