package imstm

// ctxSendRecv is the context structure for send only protocol
type ctxSendRecv struct {
	ctx    *Context
	ackReq bool //acknowledgement required?
}

// Send sends the ims message with all the message segments
func (s *ctxSendRecv) Send(segments [][]byte, ascii bool) error {
	err := send(s.ctx, segments, ascii)
	return err
}

// Recv fetches the response back
func (s *ctxSendRecv) Recv() (*Response, error) {
	resp := recv(s.ctx)
	return resp, nil
}

// Ack acknowleges the response positively
func (s *ctxSendRecv) Ack() error {
	return ack(s.ctx)
}

// Nak acknowleges the response negatively
func (s *ctxSendRecv) Nak(reason uint16, retainMsg bool) error {
	return nak(s.ctx, reason, retainMsg)
}

// WithSendRecv switches the current context into send-recv mode.
//
// checkAck, for CM1 (send-then-commit) indicates IMS to hold off the sync point commits,
// till it gets acknowledgement from IMS connect.
// checkAck, for CM0 (commit-then-send) indicates OTMA to remove the message from the output queue
//
// withTpipe indictes the CM0 communication mode
//
// purgeUndelivered, for CM0 and CM1 indicates to purge the undelivered CM0
// output messages from the tpipe message queue.
func (ctx *Context) WithSendRecv(checkAck bool, withTpipe bool, purgeUndelivered bool) SendReceiver {
	sendrecv := &ctxSendRecv{}

	//initialize irm
	ctx.irm = (&IRMHeader{}).init()
	//add data store
	copy(ctx.irm.DestID[:], A2E([]byte(ctx.session.DataStore))) //8-bytes datastore

	sendrecv.ctx = ctx

	irm := ctx.irm
	//user portion of the irm header
	irm.F1 = IRMF1CIDREQ | IRMF1MFSREQ //get client-id, modname by default
	irm.F2 = IRMF2CM1                  //if no tpipe, use CM1

	//applies to CM0
	if withTpipe {
		irm.F2 = IRMF2CM0                          //commit then send
		irm.F3 = irm.F3 | IRMF3IPURG | IRMF3SYNCNF //VERIFY: cancel the duplicate client-id
	}

	//applies to both CM0 and CM1
	if checkAck {
		irm.F3 = irm.F3 | IRMF3SYNCNF
	}

	//applies to both CM0 and CM1
	if purgeUndelivered {
		irm.F3 = irm.F3 | IRMF3PURGE
	}

	irm.F4 = IRMF4SENDRECV //send-recv protocol

	return sendrecv
}
