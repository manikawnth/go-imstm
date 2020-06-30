/*
Package imstm provides IMS connect tcp/ip client implementation.
This library uses HWSSMPL1 as the standard messaging exit.

The low level primitives include Request, Response, IRMHeader and
the higher level constructs include Session, Context, Sender, Receiver etc.
There are also helper functions to convert ascii to ebcdic and vice-versa.

While the IMS connect communication can be done by constructing the IRMHeader
and formatting the Request message structure and Parsing the Response structure,
for the most of the usecases, using Session and Context should be sufficient.

A Session represents the connection information to IMS connect:

	sess := &ims.Session{
			Addr:         "10.1.2.3:4567",
			DataStore:    "PRODIMSA",
			TLSConfig:    nil,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}

	sess.Start()

Context represents the higher level IMS connect communication protocol.
One would create a blank context sets its configuration as follows:

	//create a context from the session
	ctx := ims.NewContext(sess)

	//set the parameters on the context
	ctx.SetTranCode("ORDERTXN")

	//racf credentials can be set as follows
	ctx.SetCredentials("USER1234", "GRP123", "PASS1234")

	//setting configuration can also be chained
	ctx.SetClientID("CLIENT01").SetReroute("CLNDLQ01")

Context parameters can be set during any time of the communication and the
configuration immediately takes effect. However Only a single context can be
active at a time for a given session.

Communication with IMS connect can be started only by switching context.
Context has implementations for different IMS connect communication protocols:

1. Send-Receive using CM0 (commit-then-send) and CM1 (send-then-commit) protocol
to deal with conversational and non-conversational transactions,

	// the WithSendRecv context returns a sender-receiver interface
	sr := ctx.WithSendRecv(false, false, false)

	// create the message segments
	seg1 := []byte("ORDERTXN ITEM:GarminGPS;COUNT:2")
	seg2 := []byte("ITEM:EffectiveGo;COUNT:1")
	msg := [][]byte{seg1, seg2}

	// send the message, also indicate the converstion from ascii to ebcdic is needed
	sr.Send(msg, true)

	//receive the repsponse and print it
	resp, err := sr.Recv()
	if err != nil {
		panic(err) //handle error
	}

	//resp.Out(true) fetches the message output and optionally converts to ascii
	if outSegs, err := resp.Out(true); err !=nil {
		//handle error
	}else {
		for _, respSeg := range outSegs {
			fmt.Println(string(respSeg))
		}
	}

2. Sendonly protocol to continuously send messages to a non-response mode transaction.
Outupt of the messages sent to a response mode transactions will end up in the tpipe
queue of the same name or on the tpipe with reroute clientID set.

In the signature WithSendOnly(checkAck bool, serialDelivery bool), checkAck implies send-only
request, requesting acknowledgement from IMS connect. serialDelivery denotes that the
IMS transaction schedule type is Serial.

	//switch to send only context
	sender := ctx.WithSendOnly(false, false)

	msg := [][]byte{[]byte("NOTFYTXN ORDERID:12345, CUSTOMER:12345")}
	if err := tran.Send(msg, true); err != nil {
		//handle error
	}

3. Recvonly protocol or resume tpipe protocol to receive messages from asynchronous tpipe queue.

In the signature WithRecvOnly(singleMsg bool, flow bool, wait bool), singleMsg returns one message
at a time, where as flow implies the stream of messages from tpipe. wait parameter indicates
whether IMS connect has to wait for the messages if the tpipe queue becomes empty.
OTMA mandates the acknowledgement for such Async receives from tpipe

	//switch to recv only context
	receiver := ctx.WithRecvOnly(false, false, false)

	//fetch the response message 1
	var resp *Response
	var err error
	//receive the message
	if resp, err = Recv(); err != nil {
		panic(err) //handle error
	}

	//read the response
	if out, err := resp.Output(true); err != nil {
		//handle error
		//negative ack with reason code 35, requesting to retain the message on tpipe
		receiver.Nak(35, true)
	} else {
		for _, respSeg := range outSegs {
			fmt.Println(string(respSeg))
		}
		//acknowledge the message so that it's removed from the queue
		receiver.Ack()
	}

Please check the individual struct types for additional documentation
*/
package imstm
