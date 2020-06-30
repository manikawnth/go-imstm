package imstm

import (
	"encoding/binary"
)

// IRMHeader represents the  header portion of IMS Request Message prefix.
// It contains the total length of the message, fixed portion of the IRM header and
// the user defined portion of the IRM header as defined by HWSSMPL0/HWSSMPL1 exit message routines.
// Note: Currently correlation token is not included as synchronous callout processing is not yet
// supported
type IRMHeader struct {
	//fixed portion of the irm header
	TotLength      [4]byte //Total length of the message
	Length         [2]byte //IRMlen represents the length of the header. this should be big-endian
	Arch           byte    //irm architecture type
	F0             byte    //communication flags
	IrmID          [8]byte //irm identifier
	NakRsn         [2]byte //optional NAK reason code
	_res1          [2]byte //reserved
	F5             byte    //input message type
	Timeout        byte    //delay wait timer between ims connect and ims
	ConnType       byte    //socket connection type
	EncodingScheme byte    //unicode encoding scheme
	ClientID       [8]byte //client id for ims connect to track

	//user portion of the header. only HWSSMPL1 is supported
	F1          byte    //serves multiple purposes refer to IRMF1 constants
	F2          byte    //commit mode of the request
	F3          byte    //sync-level, routing and delivery option flags
	F4          byte    //message type
	TranCode    [8]byte //ims transaction code
	DestID      [8]byte //ims datastore name
	Lterm       [8]byte //ims lterm name
	Userid      [8]byte //racf user id
	Grpid       [8]byte //racf group id
	Passwd      [8]byte //racf password
	AppName     [8]byte //racf application name
	RerouteName [8]byte //reroute tpipe name or alternate client-id for resume tpipe call
	TagAdapt    [8]byte //name of the adapter that IMS connect calls to convert XML
	TagMap      [8]byte //name of the converter that XML adapter calls to convert XML
	ModName     [8]byte //MFS modname for input message

	//correlation token details for synchronous callout messages
	//currently not supported, hence not exported.
	ctLen   [2]byte //correlation token length
	_ctRes1 [2]byte //reserved
	imsID   [4]byte //IMS system id
	memTk   [8]byte //OTMA tmember token
	aweTk   [8]byte //OTMA message token
	ctTpipe [8]byte //OTMA tpipe name
	ctUser  [8]byte //user-id specified in ICAL call
	sesTkn  [8]byte //session value used for ims connect to ims connect connections
	extnOff [2]byte //offset value from thes start of IRM to the first IRM extension
	_ctRes2 [2]byte //reserved
}

// MarshalBinary implements BinaryMarshaler interface to encode the IRM header into byte slice
func (irm *IRMHeader) MarshalBinary() ([]byte, error) {
	len := binary.BigEndian.Uint16(irm.Length[:]) + 4
	var maxLen = 124
	out := make([]byte, maxLen)

	//fixed header
	copy(out[4:6], irm.Length[:])
	out[6] = irm.Arch
	out[7] = irm.F0
	copy(out[8:8+8], irm.IrmID[:])
	copy(out[16:16+2], irm.NakRsn[:])
	copy(out[18:18+2], irm._res1[:])
	out[20] = irm.F5
	out[21] = irm.Timeout
	out[22] = irm.ConnType
	out[23] = irm.EncodingScheme
	copy(out[24:24+8], irm.ClientID[:])

	//user portion
	out[32] = irm.F1
	out[33] = irm.F2
	out[34] = irm.F3
	out[35] = irm.F4
	copy(out[36:36+8], irm.TranCode[:])
	copy(out[44:44+8], irm.DestID[:])
	copy(out[52:52+8], irm.Lterm[:])
	copy(out[60:60+8], irm.Userid[:])
	copy(out[68:68+8], irm.Grpid[:])
	copy(out[76:76+8], irm.Passwd[:])
	copy(out[84:84+8], irm.AppName[:])
	copy(out[92:92+8], irm.RerouteName[:])
	copy(out[100:100+8], irm.TagAdapt[:])
	copy(out[108:108+8], irm.TagMap[:])
	copy(out[116:116+8], irm.ModName[:])

	if int(len) < maxLen {
		return out[:len], nil
	}
	return out, nil
}

// IRMARCH constants - architecture types
const (
	IRMARCH0 byte = iota //Y:base architectural structure for user portion
	IRMARCH1             //Y:for user portion of IRM prefix: IRM_REROUT_NM / IRM_RT_ALTCID
	IRMARCH2             //N:user portion: IRMARCH1 + IRM_TAG_ADAPT + IRM_TAG_MAP
	IRMARCH3             //N:user portion: IRMARCH2 + ICAL correlation fields + IRM_MODNAME for MFS
	IRMARCH4             //N:user portion: IRMARCH3 + IRM_SESTKN (session tokens for IMS-IMS connections)
	IRMARCH5             //N:user portion: IRMARCH4 + IRM_EXTN_OFF + 2-byte reserved field
)

// IRMF0 constants - represents flags for different IMS connect communication modes
const (
	IRMF0XMLTD   byte = 1 << iota //N:request from IMS SOAP g/w, convert XML which has both trancode and data
	IRMF0XMLD                     //N:request from IMS SOAP g/w, convert XML which has just data
	IRMF0EXTENS                   //N:message contains one or more IRM extensions
	_                             //N:'\x08' not implemented
	IRMF0NAKRSN                   //Y:NAK message with a reason code
	IRMF0SYNCNAK                  //Y:NAK message but retain the message on TPIPE queue
	IRMF0SYNASYN                  //Y:resume tpipe which fetches both sync ICAL and async message from TPIPE
	IRMF0SYNONLY                  //Y:resume tpipe only fetches ICAL sync messages
)

// IRMF5 represents flags for input message type
const (
	IRMF5SNGLNWT  byte = 1 << iota //return single message as part of resume tpipe call. doesn't wait if no message
	IRMF5AUTOFLOW                  //auto flow of the current messages one at a time. IRM_TIMER is reset after ACK of every message
	IRMF5NAUTFLOW                  //auto flow, but one at atime. IRM_TIMER causes last receive by client to terminate
	IRMF5XID                       //message includes an X/Open identifier
	IRMF5SNGLWT                    //return single message as part of resume tpipe call. waits for message
	_                              //reserved
	IRMF5NTRNSL                    //message translation done by client
	IRMF5NOTMA                     //otma headers built by client
)

// IRMSOCK represents socket connection type. Default '\x00' is transaction socket
const (
	IRMSOCKTX byte = '\x00' //Transaction socket type
	IRMSOCKP  byte = '\x10' //persistent socket type
	IRMSOCKNP byte = '\x40' //non persistent socket type. connection lasts for single exchange of input and output
)

// IRMES represents encoding scheme
const (
	IRMESUTF8  byte = '\x01' //N:UTF-8 encoding scheme
	IRMESUTF16 byte = '\x02' //N:UTF-16 encoding scheme
)

//TODO: IRM header for really user-defined portion

//TODO: re-org IRM_MAP for IRM_XID two-phase commit protocol messages

// IRMF1 represents flags for the user portion of flag F1
const (
	IRMF1TRNEXP byte = 1 << iota //expiration time for tx is set by IMS connect
	IRMF1NOWAIT                  //CM0 send-and-receive message uses NOWAIT option for the expected ACK or NAK response
	IRMF1SOARSP                  //For sendonly ACK requests, no message "text" is returned
	IRMF1UCTC                    //represents unicode transaction code
	IRMF1UC                      //represents unicode message
	IRMF1CIDREQ                  //request IMS connect to return client-id
	IRMF1MFSREQ                  //request IMS to return MFS modname
)

// IRMF2 represents commit mode flags
const (
	IRMF2UNIQCID byte = 1 << iota //request to IMS connect to generate a unique client-id
	_                             //'\x02' reserved
	_                             //'\x04' reserved
	_                             //'\x08' reserved
	_                             //'\x10' reserved
	IRMF2CM1                      //commit-mode 1 send-then-commit
	IRMF2CM0                      //commit-mode 0 commit-then-send
)

// IRMF3 represents flags for sync level, routing and delivery options for CM0 output
const (
	IRMF3SYNCNF  byte = 1 << iota //Sync level is CONFIRM
	IRMF3SYNCPT                   //Sync level is SYNCPT
	IRMF3PURGE                    //purge undeliverable CM0 output
	IRMF3REROUT                   //re-route undeliverable CM0 output
	IRMF3ORDER                    //send-only with serial delivery
	IRMF3IPURG                    //OTMA to ignore DL/I purge call for multi-segment CM0 output messages
	IRMF3DFS2082                  //for CM0 even for non-resp mode txns, issue DFS2082 if the app doesn't reply on IOPCB
	IRMF3CANCID                   //terminate an already existing session incase of duplicate client-id
)

// IRMF4 represents message type flags
const (
	IRMF4SENDRECV byte = '\x40' //' '-a send-receive transaction
	IRMF4ACK      byte = '\xC1' //'A'-an ACK response to output received from IMS connect
	IRMF4CANTIMER byte = '\xC3' //'C'-canel IRM timer for another session that has same client-id
	IRMF4DEALLOC  byte = '\xC4' //'D'-request to deallocate the conversation
	IRMF4SNDONLYA byte = '\xD2' //'K'-send-only request requires an ACK response from IMS connect
	IRMF4SYNRESPA byte = '\xD3' //'L'-send-only response to IMS call-out message requires ACK from IMS connect
	IRMF4SYNRESP  byte = '\xD4' //'M'-send-only response to sync call-out message
	IRMF4NACK     byte = '\xD5' //'N'-a NAK response for either a call-out request or sync-level CONFIRM request
	IRMF4RESTPIPE byte = '\xD9' //'R'-a resume tpipe call
	IRMF4SENDONLY byte = '\xE2' //'S'-a send-only message for non-resp mode or non-conversational IMS txns
)

// this will initialize irm to default values
func (irm *IRMHeader) init() *IRMHeader {
	irm.Arch = IRMARCH0
	irm.Length = [2]byte{'\x00', '\x50'}
	copy(irm.IrmID[:], A2E([]byte("*SAMPL1*")))
	irm.F5 = IRMF5NTRNSL    //don't tralnslate to ebcidic, we're doing it
	irm.F3 = IRMF3CANCID    //cancel duplicate client id
	irm.ConnType = IRMSOCKP //persistent socket
	irm.Timeout = '\xE9'    //default timeout
	return irm
}

// this will update the total length and architecture
func (irm *IRMHeader) setReroute(id string) *IRMHeader {
	length := uint16(96)
	binary.BigEndian.PutUint16(irm.Length[:], length)
	irm.F3 = irm.F3 | IRMF3REROUT
	copy(irm.RerouteName[:], A2E([]byte(id)))
	irm.Arch = IRMARCH2
	return irm
}
