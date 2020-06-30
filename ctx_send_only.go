package imstm

// ctxSendOnly is the context structure for send only protocol
type ctxSendOnly struct {
	ctx         *Context
	ackRequired bool //acknowledgement required?
	Sender
}

// Send sends the message using sendonly protocol
func (s *ctxSendOnly) Send(segments [][]byte, ascii bool) error {
	if err := send(s.ctx, segments, ascii); err != nil {
		return err
	}
	//if check ack is set, auto acknowledge
	if s.ackRequired {
		//TODO: remove-debug
		/*resp := recv(s.ctx)
		for {
			if typ, data, err := resp.ReadNextSegment(); err != nil {
				return err
			} else {
				fmt.Printf("type is %X\n", typ)
				fmt.Printf("%X\n", data)
				if typ == RESPSEGCSM || typ == RESPSEGERR || typ == RESPSEGINV {
					break
				}
			}
		}*/
	}
	return nil
}

// WithSendOnly returns a sender interface.
//
// ackRequired indicates that acknowledgement is needed from IMS connect for this request.
//
// serialDelivery indicates the ordered scheduling of messages, when the IMS transaction
// schedule type is defined as serial. This option will not have any effect on the parallel
// schedule type transactions
func (ctx *Context) WithSendOnly(ackRequired bool, serialDelivery bool) Sender {
	sctx := &ctxSendOnly{}
	sctx.ctx = ctx

	//initialize irm
	ctx.irm = (&IRMHeader{}).init()
	//add data store
	copy(ctx.irm.DestID[:], A2E([]byte(ctx.session.DataStore))) //8-bytes datastore

	irm := ctx.irm
	//send only has to be CM0
	irm.F2 = IRMF2CM0

	//when transaction schedule type is serial
	if serialDelivery {
		irm.F3 = irm.F3 | IRMF3ORDER
	}

	irm.F4 = IRMF4SENDONLY
	if ackRequired {
		irm.F4 = irm.F4 | IRMF4SNDONLYA //VERIFY: If IRMF4SNDONLY or IRMF4SYNRESPA
		sctx.ackRequired = true
	}

	return sctx
}
