/***** BEGIN LICENSE BLOCK *****
#
# ***** END LICENSE BLOCK *****/

package arraysplitter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mozilla-services/heka/pipeline"
	"strings"
)

type ArraySplitter struct {
	delimiter         []byte
	indexByteCount    uint
	indexByteOrder    binary.ByteOrder
	lengthByteCount   uint
	lengthByteOrder   binary.ByteOrder
	checkSumByteCount uint
	index             uint
}

type ArraySplitterConfig struct {
	Delimiter         []byte `toml:"frame_header"`
	IndexByteCount    uint   `toml:"index_count"`  // should be 0, 1, 2, 4
	IndexByteOrder    string `toml:"index_order"`  // should be "bigendian" or "b", "littleendian" or "l"
	LengthByteCount   uint   `toml:"length_count"` // should be 0, 1, 2, 4
	LengthByteOrder   string `toml:"length_order"` // should be "bigendian" or "b", "littleendian" or "l"
	CheckSumByteCount uint   `toml:"checksum_count"`
	CheckSumMethod    string `toml:"checksum_method"`
}

func (a *ArraySplitter) ConfigStruct() interface{} {
	return &ArraySplitterConfig{
		Delimiter:         []byte{'\n'},
		IndexByteCount:    uint(0),
		IndexByteOrder:    "b",
		LengthByteCount:   uint(0),
		LengthByteOrder:   "b",
		CheckSumByteCount: uint(0),
	}
}

func (a *ArraySplitter) Init(config interface{}) error {
	conf := config.(*ArraySplitterConfig)
	if len(conf.Delimiter) == 0 {
		return errors.New("ArraySplitter delimiter(frame_header) must be at least one byte.")
	}
	a.delimiter = conf.Delimiter

	if conf.IndexByteCount == 0 || conf.IndexByteCount == 1 || conf.IndexByteCount == 2 || conf.IndexByteCount == 4 {
		a.indexByteCount = conf.IndexByteCount
	} else {
		return fmt.Errorf("index_count value expect 0, 1, 2, 4. get %d", conf.IndexByteCount)
	}
	if a.indexByteCount > 0 {
		idxOrder := strings.ToLower(conf.IndexByteOrder)
		if idxOrder == "b" || idxOrder == "big" || idxOrder == "bigendian" {
			a.indexByteOrder = binary.BigEndian
		} else if idxOrder == "l" || idxOrder == "little" || idxOrder == "littleendian" {
			a.indexByteOrder = binary.LittleEndian
		} else {
			return fmt.Errorf("index_order: %s is not supported", conf.IndexByteOrder)
		}
	}

	if conf.LengthByteCount == 0 || conf.LengthByteCount == 1 || conf.LengthByteCount == 2 || conf.LengthByteCount == 4 {
		a.lengthByteCount = conf.LengthByteCount
	} else {
		return fmt.Errorf("length_count value expect 0, 1, 2, 4. get %d", conf.LengthByteCount)
	}
	if a.lengthByteCount > 0 {
		lenOrder := strings.ToLower(conf.LengthByteOrder)
		if lenOrder == "b" || lenOrder == "big" || lenOrder == "bigendian" {
			a.lengthByteOrder = binary.BigEndian
		} else if lenOrder == "l" || lenOrder == "little" || lenOrder == "littleendian" {
			a.lengthByteOrder = binary.LittleEndian
		} else {
			return fmt.Errorf("length_order: %s is not supported", conf.LengthByteOrder)
		}
	}

	a.checkSumByteCount = conf.CheckSumByteCount
	return nil
}

func (a *ArraySplitter) FindRecord(buf []byte) (bytesRead int, record []byte) {
	result := bytes.SplitN(buf, a.delimiter, 3)
	if len(result) < 3 {
		return 0, nil
	}
	// Include the first delimiter in what's been read.
	// Include the item before the first delimiter, in normal len(result[0]) should be 0
	// Not include the second delimiter
	bytesRead = len(result[0]) + len(a.delimiter) + len(result[1])

	ol := int(a.indexByteCount + a.lengthByteCount + a.checkSumByteCount)
	if len(result[1]) < ol {
		fmt.Println("Splitter error: frame data length too short")
		return bytesRead, nil // TODO: bad frame, through away, and make some mark
	}

	data := result[1]
	b := bytes.NewBuffer(data)
	switch a.indexByteCount {
	case 1:
		index := uint8(data[0])
		if index != uint8(a.index)+1 {
			fmt.Println("Splitter: frame index increment error: ", a.index, " to ", index)
		}
		a.index = uint(index)
	case 2:
		index := uint16(0)
		binary.Read(b, a.indexByteOrder, &index)
		if index != uint16(a.index)+1 {
			fmt.Println("Splitter: frame index increment error: ", a.index, " to ", index)
		}
		a.index = uint(index)
	case 4:
		index := uint32(0)
		binary.Read(b, a.indexByteOrder, &index)
		if index != uint32(a.index)+1 {
			fmt.Println("Splitter: frame index increment error: ", a.index, " to ", index)
		}
		a.index = uint(index)
	default:
	}

	data = data[a.indexByteCount:]
	dl := len(result[1]) - ol
	switch a.lengthByteCount {
	case 1:
		length := uint8(data[0])
		if length != uint8(dl) {
			fmt.Println("Splitter: frame data length not match, expect ", length, " get ", dl)
		}
	case 2:
		length := uint16(0)
		binary.Read(b, a.lengthByteOrder, &length)
		if length != uint16(dl) {
			fmt.Println("Splitter: frame data length not match, expect ", length, " get ", dl)
		}
	case 4:
		length := uint32(0)
		binary.Read(b, a.lengthByteOrder, &length)
		if length != uint32(dl) {
			fmt.Println("Splitter: frame data length not match, expect ", length, " get ", dl)
		}
	default:
	}

	data = data[a.lengthByteCount:]
	if a.checkSumByteCount > 0 {
		cs := data[len(data)-int(a.checkSumByteCount):]
		data = data[:len(data)-int(a.checkSumByteCount)]
		fmt.Println("Splitter: frame data checksum ", cs)
	}

	// return bytesRead, buf[len(result[0]):bytesRead]
	return bytesRead, data
}

func init() {
	pipeline.RegisterPlugin("ArraySplitter", func() interface{} {
		return new(ArraySplitter)
	})
}
