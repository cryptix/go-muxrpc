package muxrpc

import (
	"context"
	"testing"

	"cryptoscope.co/go/luigi"
	"github.com/stretchr/testify/require"
)

func TestHandlerMux(t *testing.T) {
	mux := &HandlerMux{}
	r := require.New(t)
	call := make(chan struct{})
	connect := make(chan struct{})

	back1, forth1 := luigi.NewDuplexPipe()
	back2, _ := luigi.NewDuplexPipe()

	exp := &Request{
		Method: Method{"foo", "bar"},
		Stream: NewStream(back1, back1, 1, true, true),
	}
	notexp := &Request{
		Method: Method{"goo", "bar"},
		Stream: NewStream(back2, back2, 2, true, true),
	}

	handler := &testHandler{
		call: func(ctx context.Context, req *Request) {
			r.Equal(exp.Method.String(), req.Method.String(), "Method doesn't match")
			req.Stream.Close()
			close(call)
		},
		connect: func(ctx context.Context, e Endpoint) {
			close(connect)
		},
	}

	mux.Register(Method{"foo", "bar"}, handler)

	go func() {
		mux.HandleCall(context.TODO(), exp)
		mux.HandleCall(context.TODO(), notexp)

		mux.HandleConnect(context.TODO(), nil)
	}()

	for call != nil || connect != nil {
		select {
		case <-call:
			call = nil
		case <-connect:
			connect = nil
		}
	}

	_, err := forth1.Next(context.TODO())
	r.Error(luigi.EOS{}, err)
}