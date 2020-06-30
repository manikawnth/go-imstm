package imstm

// ctxRecvOnly is the context structure for recv only or resume-tpipe protocol
type ctxRecvOnly struct {
	ctx     *Context
	initial bool //identifies if the current context is in initial state
	Receiver
}

// Recv fetches the messages from async hold queue
func (r *ctxRecvOnly) Recv() (*Response, error) {
	defer func() {

	}()
	if !r.initial {
		if err := send(r.ctx, nil, false); err != nil {
			return nil, err
		}
		r.initial = true
	}
	resp := recv(r.ctx)
	return resp, nil
}

// Ack acknowleges the response positively
func (r *ctxRecvOnly) Ack() error {
	return ack(r.ctx)
}

// Nak acknowleges the response negatively
func (r *ctxRecvOnly) Nak(reason uint16, retainMsg bool) error {
	return nak(r.ctx, reason, retainMsg)
}

// WithRecvOnly returns a receiver interface by switching the context to recv messages asynchronously
// using resume tpipe protocol.
//
// singleMsg - setting to true indicates IMS connect to fetch only one message in this context.
//
// flow - setting to true indicates OTMA to fetch the messages on tpipe continuously.
//
// wait - setting to true informs IMS connect to wait on OTMA for new messages if the current messages
// are exhausted on the queue.
//
// Acknowledgement of messages, using Ack() or Nak(..) are necessary after the receipt of messages.
func (ctx *Context) WithRecvOnly(singleMsg bool, flow bool, wait bool) Receiver {
	sctx := &ctxRecvOnly{}
	sctx.ctx = ctx

	//initialize irm
	ctx.irm = (&IRMHeader{}).init()
	//add data store
	copy(ctx.irm.DestID[:], A2E([]byte(ctx.session.DataStore))) //8-bytes datastore

	irm := ctx.irm

	//for recv only
	irm.F2 = IRMF2CM0

	// Resume tpipe
	irm.F4 = irm.F4 | IRMF4RESTPIPE

	if singleMsg && wait {
		irm.F5 = irm.F5 | IRMF5SNGLWT
	} else if singleMsg && !wait {
		irm.F5 = irm.F5 | IRMF5SNGLNWT
	} else if flow && wait {
		irm.F5 = irm.F5 | IRMF5AUTOFLOW
	} else if flow && !wait {
		irm.F5 = irm.F5 | IRMF5NAUTFLOW
	}

	return sctx
}
