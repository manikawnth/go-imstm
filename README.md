# go-imstm

## Overview
This package provides a tcp/ip client interface to communicate with IMS connect to perform IMS transaction management operations

It provides interface for low-level primitives like `Request`, `Response`, `IRMHeader` 

**IRMHeader** - Provides interface to construct IMS Request Message prefix to be compatible with the HWSSMPL1 user message exit routine

**Request** - Provides interface to build a new Request message with the supplied IRM prefix, add the message segments and wrie the request on to TCP connection

**Response** - Provides interface to create a Response message structure, from which you can read individual segments from the client connection. It also contains additional IMS connect message structures like Complete Status Message (for successful completion), Request Status Message (for erroneous completion)

However, the users often just need the high-level abstractions like `Context` which provide interface to different protocols like SendReceive, SendOnly, RecvOnly etc.

**Context** - Provides interface for invoking protocols and also set the configuration. You need to switch the context to start communication using a different protocol.

## Usage

**send-receive usage:**

```go
  package main

  import (
    "fmt"
    "log"
    "time"

    "github.com/manikawnth/go-imstm"
  )

  func main() {

    sess := &imstm.Session{
      Addr:         "10.1.2.3:4567",
      DataStore:    "PRODIMSA",
      TLSConfig:    nil,
      ReadTimeout:  5 * time.Second,
      WriteTimeout: 5 * time.Second,
    }

    if err := sess.Start(); err != nil {
      panic(err)
    }
    defer sess.End()

    ctx := imstm.NewContext(sess)
    sendreceiver := ctx.WithSendRecv(false, false, false)

    ctx.SetTranCode("ORDERTXN").SetCredentials("USER1234", "GRP123", "PASS1234")
    ctx.SetClientID("CLIENT01")

    seg1 := []byte("ORDERTXN ITEM:GOPHER;COUNT:2")
    msg := [][]byte{seg1}

    //send the message in ascii and let the interface convert it
    if err := sendreceiver.Send(msg, true); err != nil {
      panic(err)
    }

    resp, err := sendreceiver.Recv()
    if err != nil {
      log.Fatalln(err)
    }
    outSegs, err := resp.Out(true)
    if err != nil {
      log.Fatalf("resp processing error: %v", err)
    }
    for _, outSeg := range outSegs {
      fmt.Println(string(outSeg))
    }
  }
```


**send-only and recv-only usage:**

```go
  package main

  import (
    "fmt"
    "log"
    "time"

    "github.com/manikawnth/go-imstm"
  )

  func main() {

    sess := &imstm.Session{
      Addr:         "10.1.2.3:4567",
      DataStore:    "PRODIMSA",
      TLSConfig:    nil,
      ReadTimeout:  5 * time.Second,
      WriteTimeout: 5 * time.Second,
    }

    if err := sess.Start(); err != nil {
      panic(err)
    }
    defer sess.End()

    ctx := imstm.NewContext(sess)

    //SWITCH TO SEND ONLY CONTEXT
    sender := ctx.WithSendOnly(false, false)

    ctx.SetTranCode("ORDERTXN").SetCredentials("USER1234", "GRP123", "PASS1234")
    ctx.SetClientID("CLIENT01")SetReroute("CLNTDLQ1")	

    //send the sendonly non-response mode txn
    msgSeg := []byte("NOTFYTXN ORDERID:12345, CUSTOMER:12345")
    msg := [][]byte{msgSeg}
    if err := sender.Send(msg, true); err != nil {
      panic(err)
    }
    //send the response mode txn
    msgSeg := []byte("ORDERTXN ITEM:GOPHER;COUNT:2")
    msg := [][]byte{msgSeg}
    if err := sender.Send(msg, true); err != nil {
      panic(err)
    }
    

    //SWITCH TO RECV ONLY CONTEXT
    receiver := ctx.WithRecvOnly(false, false, false)

    //since context is reset, set the necessary config as again
    ctx.SetCredentials("USER1234", "GRP123", "PASS1234")
    ctx.SetClientID("CLIENT01")SetReroute("CLNTDLQ1")
    //receive the async response generated from 2nd sendonly request above
    resp, err := receiver.Recv()
    if err != nil {
      log.Fatalln(err)
    }
    outSegs, err := resp.Out(true)
    if err != nil {
      log.Fatalf("resp processing error: %v", err)
    }
    for _, outSeg := range outSegs {
      fmt.Println(string(outSeg))
    }

    //acknowledge the response
    receiver.Ack()
  }
```

## Roadmap

- [ ] support for ping message and background health-check
- [ ] filling lacking IMS timeout configuration
- [ ] support for synchronous callouts
- [ ] connection pooling (little tricky from interfacing, each connection is unique client for IMS)
- [ ] dynamic client id generation
- [ ] higher level interface for Type 1 commands