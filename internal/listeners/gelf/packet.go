package gelf

import (
	"bytes"
)

type Packet struct {
	Addr string
	Data []byte
}

// chunkMagicBytes returns the magic number used to identify a GELF chunk.
func (p Packet) chunkMagicBytes() []byte {
	return []byte{0x1e, 0x0f}
}

// gzipMagicBytes returns the magic number used to identify a Gzip compressed GELF payload.
func (p Packet) gzipMagicBytes() []byte {
	return []byte{0x1f, 0x8b}
}

// zlibMagicBytes returns the magic number used to identify a Zlib compressed GELF payload.
func (p Packet) zlibMagicBytes() []byte {
	return []byte{0x78}
}

// IsChunk checks if the packet is a GELF chunk.
func (p Packet) IsChunk() bool {
	return bytes.HasPrefix(p.Data, p.chunkMagicBytes())
}

// IsGzipCompressed checks if the packet is Gzip compressed.
func (p Packet) IsGzipCompressed() bool {
	return bytes.HasPrefix(p.Data, p.gzipMagicBytes())
}

// IsZlibCompressed checks if the packet is Zlib compressed.
func (p Packet) IsZlibCompressed() bool {
	return bytes.HasPrefix(p.Data, p.zlibMagicBytes())
}
