package encoding

import (
	"bytes"
	"encoding/gob"

	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/sonic"
)

type Encoder interface {
	Encode(data any) ([]byte, error)
	Decode(encodeData []byte, decodeData any) error
}

func NewGobEncoder() Encoder {
	return &gobEncoder{}
}

type gobEncoder struct{}

func (g *gobEncoder) Encode(data any) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(data); err != nil {
		return nil, errorx.Wrapf(err, "GobEncoder.Encode failed")
	}
	return buff.Bytes(), nil
}

func (g *gobEncoder) Decode(encodeData []byte, decodeData any) error {
	buff := bytes.NewBuffer(encodeData)
	dec := gob.NewDecoder(buff)
	if err := dec.Decode(&decodeData); err != nil {
		return errorx.Wrapf(err, "GobEncoder.Decode failed")
	}

	return nil
}

type jsonEncoder struct{}

func NewJSONEncoder() Encoder {
	return &jsonEncoder{}
}

func (j *jsonEncoder) Encode(data any) ([]byte, error) {
	encodeData, err := sonic.Marshal(data)
	if err != nil {
		return nil, errorx.Wrapf(err, "JSONEncoder.Encode failed")
	}

	return encodeData, nil
}

func (j *jsonEncoder) Decode(encodeData []byte, decodeData any) error {
	if err := sonic.Unmarshal(encodeData, &decodeData); err != nil {
		return errorx.Wrapf(err, "JSONEncoder.Decode failed")
	}

	return nil
}
