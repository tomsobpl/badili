package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func DecodePacketToProtoMessage(p Packet) (*gelfapi.Message, error) {
	if p.IsChunk() {
		return nil, errors.New("GELF chunk data is not supported")
	}

	var (
		data   = p.Data
		reader io.ReadCloser
		err    error
		msg    = &gelfapi.Message{}
		rawMsg map[string]any
	)

	switch {
	case p.IsGzipCompressed():
		reader, err = gzip.NewReader(bytes.NewReader(p.Data))
	case p.IsZlibCompressed():
		reader, err = zlib.NewReader(bytes.NewReader(p.Data))
	}

	if err != nil {
		return nil, err
	}

	if reader != nil {
		defer func(reader io.ReadCloser, err error) {
			err = reader.Close()
		}(reader, err)

		if data, err = io.ReadAll(reader); err != nil {
			return nil, err
		}
	}

	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return nil, err
	}

	if v, ok := rawMsg["version"].(string); ok {
		msg.Version = v
		delete(rawMsg, "version")
	}

	if v, ok := rawMsg["host"].(string); ok {
		msg.Host = v
		delete(rawMsg, "host")
	}

	if v, ok := rawMsg["short_message"].(string); ok {
		msg.ShortMessage = v
		delete(rawMsg, "short_message")
	}

	if v, ok := rawMsg["full_message"].(string); ok {
		msg.FullMessage = v
		delete(rawMsg, "full_message")
	}

	if v, ok := rawMsg["timestamp"].(float64); ok {
		msg.Timestamp = v
		delete(rawMsg, "timestamp")
	}

	if v, ok := rawMsg["level"].(float64); ok {
		msg.Level = int32(v)
		delete(rawMsg, "level")
	}

	if len(rawMsg) > 0 {
		extras, err := structpb.NewStruct(rawMsg)
		if err != nil {
			return nil, err
		}
		msg.Extras = extras
	}

	//slog.Info("msg after unmarshal", "msg", msg)

	return msg, nil
}
