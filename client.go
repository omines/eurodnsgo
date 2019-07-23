package eurodnsgo

import (
	"context"
	"errors"
	"log"
	"time"
)

// The delay between calls in milliseconds between scheduled
// requests.
const defaultCallDelay = 500

// ClientConfig represents the data needed to connect to the API
type ClientConfig struct {
	// Set the Host to connect to to use the API
	Host string
	// Set the Username to connect to the API
	Username string
	// Set the Password to connect to the API
	Password string
	// The CallDelay regulates the schedule iteration speed
	// in milliseconds. Defaults to 500 milliseconds.
	CallDelay int
}

// Client defines the functions needed to do a remote request
type Client interface {
	// Schedule schedules a request to be send to the XML server
	Schedule(context.Context, *SoapRequest) (chan []byte, error)
	// Call performs a request to the XML server
	Call(context.Context, *SoapRequest) error
}

type scheduledCall struct {
	sr     *SoapRequest
	result chan []byte
	// Context is required in this struct since it is passed through a channel
	ctx context.Context
}

type client struct {
	sc           *soapClient
	callDelay    int
	callSchedule chan scheduledCall
}

// Schedule schedules a request to be send to the EuroDNS server
func (c *client) Schedule(ctx context.Context, sr *SoapRequest) (chan []byte, error) {
	r := make(chan []byte, 1)
	// will be processed inside client::run
	c.callSchedule <- scheduledCall{sr, r, ctx}
	return r, nil
}

// Call performs a request at the EuroDNS server directly. The use
// of Schedule is advised to prevent flooding the server with
// requests
func (c *client) Call(ctx context.Context, sr *SoapRequest) error {
	_, err := c.makeCall(ctx, sr)
	return err
}

func (c *client) run() {
	for {
		select {
		case sc := <-c.callSchedule:
			b, err := c.makeCall(sc.ctx, sc.sr)
			if err != nil {
				log.Println("Error calling API")
				log.Println(err)
				sc.result <- []byte{}
				break
			}
			sc.result <- b
			time.Sleep(time.Duration(c.callDelay) * time.Millisecond)
		default:
			// cap the process
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (c *client) makeCall(ctx context.Context, sr *SoapRequest) ([]byte, error) {
	return c.sc.call(ctx, sr)
}

// NewClient returns a new client with the appropriate credentials
// setup.
func NewClient(cc ClientConfig) (Client, error) {
	if len(cc.Username) == 0 {
		return nil, errors.New("A username should be provided")
	}

	if len(cc.Password) == 0 {
		return nil, errors.New("A password should be provided")
	}

	if len(cc.Host) == 0 {
		return nil, errors.New("A host URL should be provided")
	}

	// callDelay in milliseconds between scheduled calls
	callDelay := defaultCallDelay
	if cc.CallDelay > 0 {
		callDelay = cc.CallDelay
	}

	sc := &soapClient{
		cc.Username,
		cc.Password,
		cc.Host,
		callDelay,
	}

	c := &client{
		sc:           sc,
		callSchedule: make(chan scheduledCall, 32),
	}
	go c.run()

	return c, nil
}
