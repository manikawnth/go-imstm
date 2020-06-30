package imstm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// ErrInvalidUnmarshal indicates an error while unmarshaling the response segment
var ErrInvalidUnmarshal = errors.New("Invalid Unmarshal Object")

// ErrSegmentNotPresent indicates the requested segment not present in the response message
var ErrSegmentNotPresent = errors.New("Segment not present")

// RespRMM represents Request Mod Message in the response
type RespRMM struct {
	LL  [2]byte
	ZZ  [2]byte
	ID  [8]byte
	MOD [8]byte
}

// UnmarshalBinary unmarshals the segment into RMM structure
func (rmm *RespRMM) UnmarshalBinary(data []byte) error {
	if len(data) >= 20 {
		copy(rmm.LL[:], data[:2])
		copy(rmm.ZZ[:], data[2:2+2])
		copy(rmm.ID[:], data[4:4+8])
		copy(rmm.MOD[:], data[12:12+8])
		return nil
	}
	return ErrInvalidUnmarshal
}

// RespCID represents Request Client Id in the response
type RespCID struct {
	LL       [2]byte
	ZZ       [2]byte
	ID       [8]byte
	ClientID [8]byte
}

// UnmarshalBinary unmarshals the segment into CID structure
func (cid *RespCID) UnmarshalBinary(data []byte) error {
	if len(data) >= 20 {
		copy(cid.LL[:], data[:2])
		copy(cid.ZZ[:], data[2:2+2])
		copy(cid.ID[:], data[4:4+8])
		copy(cid.ClientID[:], data[12:12+8])
		return nil
	}
	return ErrInvalidUnmarshal
}

// RespCSM represents Complete Status Message in the response
type RespCSM struct {
	LL        [2]byte
	MsgFlag   byte
	ProtoFlag byte
	ID        [8]byte
}

// UnmarshalBinary unmarshals the segment into CSM structure
func (csm *RespCSM) UnmarshalBinary(data []byte) error {
	if len(data) >= 12 {
		copy(csm.LL[:], data[:2])
		csm.MsgFlag = data[2]
		csm.ProtoFlag = data[3]
		copy(csm.ID[:], data[4:4+8])
		return nil
	}
	return ErrInvalidUnmarshal
}

// RespRSM represents Request Status Message in the response - sent only incase of errors
type RespRSM struct {
	LL         [2]byte //Length of the Request status message
	StatusFlag byte    //status flag
	RacfRc     byte    //RACF Reason code for security errors
	ID         [8]byte //Id - will have *REQSTS*
	RetCode    [4]byte //Return code
	RsnCode    [4]byte //Reason code
}

// UnmarshalBinary unmarshals the segment into RSM structure
func (rsm *RespRSM) UnmarshalBinary(data []byte) error {
	if len(data) >= 20 {
		copy(rsm.LL[:], data[:2])
		rsm.StatusFlag = data[2]
		rsm.RacfRc = data[3]
		copy(rsm.ID[:], data[4:4+8])
		copy(rsm.RetCode[:], data[12:12+4])
		copy(rsm.RsnCode[:], data[16:16+4])
		return nil
	}
	return ErrInvalidUnmarshal
}

//TODO: RespCT is the correlation token structure for synchronous call-out requests from IMS
type respCT struct {
	Length [2]byte //Total length of the strcuture including the token
	_res   [2]byte
	ID     [8]byte //contains *CORTKN* as identifier
	LL     [2]byte
	_res1  [2]byte
	ImsID  [4]byte //IMS system id
	memTk  [8]byte //OTMA Tmember token
	AweTk  [8]byte //OTMA Message token
	Tpipe  [8]byte //TPIPE name
	UserID [8]byte //userid included in the ICAL call by IMS application
}

// respPing is the Ping response from IMS connect
type respPing struct {
	LL   [2]byte  //length of ping response
	F1   byte     //'\x01'support for IMS extensions present
	_F2  byte     //flag reserved
	Resp [25]byte //Maximum of 25 byte response which contains '*PING RESPONSE*' or 'HWSC0030I *PING RESPONSE*'
}

// respDataSeg represents the message segment
type respDataSeg struct {
	LL   [2]byte
	ZZ   [2]byte
	Data []byte
}

// Response represents the IMS connect response message
type Response struct {
	length  uint32        //total length of the response
	reader  io.Reader     //reader stored here
	timeout time.Duration //timeout in ms to fetch each segment
	initial bool          //at the start of the message?
	retCode uint32        //ims connect return code
	rsnCode uint32        //ims connect reason code
	rmm     []byte        //request mod message
	cid     []byte        //client-id message
	csm     []byte        //complete status message. marks success
	rsm     []byte        //request status message, marks error
	cortok  []byte        //correlation token for sync callouts
	data    [][]byte      //data segments
}

// RespSegType is the type of segment in the IMS connect response
type RespSegType byte

// List of Response segment types
const (
	RESPSEGINV  RespSegType = '\x00' //Invalid segment
	RESPSEGERR  RespSegType = 'E'    //Error Segment - *REQSTS*
	RESPSEGRMM  RespSegType = 'M'    //RMM segment - *REQMOD*
	RESPSEGCID  RespSegType = 'C'    //CID segment - *GENCID*
	RESPSEGCT   RespSegType = 'T'    //Correlation Token segment - *CORTKN*
	RESPSEGDATA RespSegType = 'D'    //Data segment
	RESPSEGCSM  RespSegType = 'S'    //Status CSM segment - *CSMOKY*
)

// ReadNextSegment fetches the next segment.
// Internally a net.Conn read call is issued to fetch the next segment of the message.
// It retursn the type of the segment, data of the segment and an error.
//
// Error is generally a net.Conn read error, like a read timeout or buffer length errors
// if IMS connect populates the length in the response incorrectly.
//
// Typically users invoke the Out() method of Response to read all the segments of the response.
// When invoking this method, you should check for the segment types RESPSEGINV - which represents
// invalid segment, RESPSEGERR - which contains Request Status Message indicating an error in the
// IMS connect response, RESPSEGCSM - which contains Complete Status Message indicating successful
// response.
func (r *Response) ReadNextSegment() (segType RespSegType, segData []byte, err error) {

	segType = RESPSEGINV
	//reset the read timeout at the once the complete message is read
	defer func() {
		switch segType {
		case RESPSEGINV:
		case RESPSEGERR:
		case RESPSEGCSM:
			r.reader.(net.Conn).SetReadDeadline(time.Now().Add(0 * time.Second))
		}
	}()

	var segLen uint16
	var length [4]byte

	//get the total length
	if r.initial {
		r.initial = false
		r.reader.(net.Conn).SetReadDeadline(time.Now().Add(r.timeout))
		if _, err = io.ReadFull(r.reader, length[:]); err != nil {
			goto badExit
		}
		r.length = binary.BigEndian.Uint32(length[:4])
	}

	// read each segment
	if _, err = io.ReadFull(r.reader, length[:2]); err != nil {
		goto badExit
	}
	//TODO: n can't be less than 2
	segLen = binary.BigEndian.Uint16(length[:2])
	segData = make([]byte, int(segLen))
	copy(segData[:2], length[:2])
	if _, err = io.ReadFull(r.reader, segData[2:]); err != nil {
		goto badExit
	}

	segType = RESPSEGDATA
	if segLen >= 12 {
		switch string(E2A(segData[4:12])) {
		case "*REQSTS*":
			segType = RESPSEGERR
			break
		case "*REQMOD*":
			segType = RESPSEGERR
			break
		case "*GENCID*":
			segType = RESPSEGCID
			break
		case "*CSMOKY*":
			segType = RESPSEGCSM
			break
		case "*CORTOKN":
			segType = RESPSEGCT
			break
		}
	}
	goto goodExit
badExit:
	//TODO: wrap the errors in future
goodExit:
	return segType, segData, err
}

// readAllSegments reads all the segments in the output message at once
func (r *Response) readAllSegments() error {
	var end bool
	for {
		segType, segData, err := r.ReadNextSegment()
		if err != nil {
			return err
		}
		switch segType {
		case RESPSEGERR:
			r.rsm = segData
			var rsm RespRSM
			(&rsm).UnmarshalBinary(r.rsm)
			r.retCode = binary.BigEndian.Uint32(rsm.RetCode[:])
			r.rsnCode = binary.BigEndian.Uint32(rsm.RsnCode[:])
			end = true
			break
		case RESPSEGCT:
			r.cortok = segData
			break
		case RESPSEGRMM:
			r.rmm = segData
			break
		case RESPSEGCID:
			r.cid = segData
			break
		case RESPSEGDATA:
			r.data = append(r.data, segData)
			break
		case RESPSEGCSM:
			r.csm = segData
			end = true
			break
		}
		if end {
			return nil
		}
	}
}

// Out will return the complete response message from IMS.
// Passing ascii parameter as true converts each byte slice into ascii from ebcidic.
// This may not be useful when your response message is not completely text data.
//
// Error can represent the network errors, io read errors and more important the
// IMS connect return and reason codes present in the RSM segment, like security violations etc.
func (r *Response) Out(ascii bool) ([][]byte, error) {
	var err error
	//read all the segments of the message
	if err = r.readAllSegments(); err != nil {
		return nil, err
	}

	//if error segment present, nothing else can exist
	if r.rsm != nil {
		err = fmt.Errorf("ErrIMSConnect: ReturnCode: %d, ReasonCode: %d", r.retCode, r.rsnCode)
		return nil, err
	}
	var out [][]byte
	//if we have csm, then there's a definite output
	if r.csm != nil {
		// if no data, just return back
		if r.data == nil {
			return nil, nil //TODO: should we rather throw "no output" error?
		}
		//loop over the data
		for _, seg := range r.data {
			if ascii {
				out = append(out, E2A(seg[4:]))
			} else {
				segCopy := make([]byte, len(seg)-4)
				out = append(out, segCopy[4:])
			}
		}
		return out, nil
	}
	return nil, ErrSegmentNotPresent
}

// ModName returns the modname from the IOPCB ISRT call
func (r *Response) ModName() (string, error) {
	if r.rmm == nil {
		return "", ErrSegmentNotPresent
	}
	rmm := &RespRMM{}
	rmm.UnmarshalBinary(r.rmm)
	return string(E2A(rmm.MOD[:])), nil
}

// ClientID returns any clientid that is generated by IMS connect
func (r *Response) ClientID() (string, error) {
	if r.cid == nil {
		return "", ErrSegmentNotPresent
	}
	cid := &RespCID{}
	cid.UnmarshalBinary(r.cid)
	return string(E2A(cid.ClientID[:])), nil
}

// NewResponse returns the new initialized response structure.
// This is typically returned from context's receive call and the users
// invoke Out() or ReadNextSegment() of the Response.
func NewResponse(reader io.Reader, timeout time.Duration) *Response {
	var r Response
	r.initial = true
	r.timeout = timeout
	r.reader = reader
	return &r
}
