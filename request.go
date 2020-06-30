package imstm

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

// Request represents the IMS connect request
type Request struct {
	length    uint32
	writer    io.Writer     //writer stored here
	timeout   time.Duration //timeout in ms to fetch each segment
	irmHeader *IRMHeader    //irm header for the request, owned by request
	segments  [][]byte      //message segments, owned by request
}

const maxSegLen = 32 * 1024

// RequestTrailer marks end of the request for HWSSMPL0/HWSSMPL1 message exit routines
var RequestTrailer = []byte{'\x00', '\x04', '\x00', '\x00'}

// AddSegment will add segment to the the request buffer
// Users do not need to provide the LL and ZZ.
// Length is automatically inferred from the length of the input byte slice
func (r *Request) AddSegment(segment []byte) *Request {
	segLen := len(segment)
	if segLen > maxSegLen-4 {
		segLen = maxSegLen - 4
	}
	seg := make([]byte, segLen+4)
	binary.BigEndian.PutUint16(seg[:2], uint16(segLen+4)) //LL, ZZ is already initialized to 0
	copy(seg[4:], segment[:])
	r.segments = append(r.segments, seg)
	r.length = r.length + uint32(len(seg))
	return r
}

// Write writes the request message on to the network connection writer interface
func (r *Request) Write() error {
	defer r.writer.(net.Conn).SetWriteDeadline(time.Now().Add(0 * time.Second))
	header, _ := r.irmHeader.MarshalBinary()
	//populate the total length
	binary.BigEndian.PutUint32(header[:4], r.length)
	r.writer.(net.Conn).SetWriteDeadline(time.Now().Add(r.timeout))
	//TODO: handle errors and incomplete writes, although not an issue for sockets
	//write the header
	if _, err := r.writer.Write(header); err != nil {
		return err
	}
	for _, seg := range r.segments {
		if _, err := r.writer.Write(seg); err != nil {
			return err
		}
	}
	_, err := r.writer.Write(RequestTrailer)
	return err
}

// NewRequest function creates a new requtest with the supplied IRM header and write timeout
// parameters. The request message is automatically constructed when the users invoke the
// corresponding Context's Send() method
func NewRequest(writer io.Writer, irmHeader IRMHeader, timeout time.Duration) *Request {
	var r Request
	r.irmHeader = &irmHeader
	r.writer = writer
	r.timeout = timeout
	r.length = 4 + uint32(binary.BigEndian.Uint16(r.irmHeader.Length[:])) + 4
	return &r
}
