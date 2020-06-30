package imstm

import (
	"encoding/binary"
	"time"
)

// Sender sends the IMS connect message
type Sender interface {
	Send(segments [][]byte, ascii bool) error
}

// Receiver receives the IMS connect response
type Receiver interface {
	Recv() (*Response, error)
	Acknowledger
}

// Acknowledger acknowledges the messages after receipt
type Acknowledger interface {
	Ack() error
	Nak(reason uint16, retainMsg bool) error
}

// SendReceiver sends and receives the messages
type SendReceiver interface {
	Sender
	Receiver
}

// Context is a structure that holds the connection and state details of the IMS connect communication Context
type Context struct {
	session *Session
	irm     *IRMHeader
	active  bool //tells if a context is already active
}

// SetReroute adds the client id to the irm header
func (ctx *Context) SetReroute(clientID string) *Context {
	ctx.irm.setReroute(clientID)
	return ctx
}

// SetClientID adds the client id to the irm header
func (ctx *Context) SetClientID(clientID string) *Context {
	copy(ctx.irm.ClientID[:], A2E([]byte(clientID))) //8-bytes client id
	return ctx
}

// SetTranCode adds the transaction id to the irm header
func (ctx *Context) SetTranCode(tranCode string) *Context {
	copy(ctx.irm.TranCode[:], A2E([]byte(tranCode))) //8-bytes transaction code
	return ctx
}

// SetLterm adds the lterm override to the iopcb
func (ctx *Context) SetLterm(lterm string) *Context {
	copy(ctx.irm.Lterm[:], A2E([]byte(lterm))) //8-bytes ltermoverride
	return ctx
}

// SetModName adds the mod name to the iopcb
func (ctx *Context) SetModName(modName string) *Context {
	copy(ctx.irm.ModName[:], A2E([]byte(modName))) //8-bytes mod name
	return ctx
}

// SetCredentials adds the racf credentials
func (ctx *Context) SetCredentials(userid string, grpid string, passwd string) *Context {
	copy(ctx.irm.Userid[:], A2E([]byte(userid)))
	copy(ctx.irm.Grpid[:], A2E([]byte(grpid)))
	copy(ctx.irm.Passwd[:], A2E([]byte(passwd)))
	return ctx
}

// TODO: for irm timer - setTimeout adds the lterm override to the iopcb
func (ctx *Context) setTimeout(timeout time.Duration) *Context {
	return ctx
}

// send sends a message with multiple segments
func send(ctx *Context, segments [][]byte, ascii bool) error {
	request := NewRequest(ctx.session.conn, *ctx.irm, ctx.session.WriteTimeout)
	for _, segment := range segments {
		if ascii {
			request.AddSegment(A2E(segment))
		}
	}

	return request.Write()
}

// recv receives a response message
func recv(ctx *Context) *Response {
	return NewResponse(ctx.session.conn, ctx.session.ReadTimeout)
}

// ack acknowledges positively
func ack(ctx *Context) error {
	oldF4 := ctx.irm.F4
	defer func() {
		ctx.irm.F4 = oldF4
	}()
	ctx.irm.F4 = IRMF4ACK
	return send(ctx, nil, false)
}

// nak acknowledges negatively
func nak(ctx *Context, reason uint16, retainMsg bool) error {
	oldF0 := ctx.irm.F0
	oldNakRsn := ctx.irm.NakRsn
	oldF4 := ctx.irm.F4
	defer func() {
		ctx.irm.F0 = oldF0
		ctx.irm.NakRsn = oldNakRsn //check the beauty of go arrays which perform value copy
		ctx.irm.F4 = oldF4
	}()
	ctx.irm.F4 = IRMF4NACK //negative ack
	if retainMsg {
		ctx.irm.F0 = IRMF0SYNCNAK //keep the message on tpipe queue
	}
	if reason != 0 {
		ctx.irm.F0 = ctx.irm.F0 | IRMF0NAKRSN
		binary.BigEndian.PutUint16(ctx.irm.NakRsn[:], reason)
	}
	return send(ctx, nil, false)
}

// end - ends the Context but doesn't close the underlying connection
func (ctx *Context) end() error {
	return nil
}

// NewContext creates and returns a new context
func NewContext(session *Session) *Context {
	ctx := &Context{}
	ctx.session = session
	return ctx
}
