package aws

import (
	"errors"
	"net"
	"sync"
	"time"
	//  "log"
)

// Dev notes: lower-case (private) functions assume the lock is held,
// upper-case functions should use a defer lock.Unlock to ensure
// underlying dialer/socket panics will not leave locks hanging.

var ErrUnderlyingNotconnected = errors.New("Underlying socket is not connected")

// A Dialer is usually a closuer that
// is pre-configured to the callers tastes.
//
// (see URLDialer for an example/default generator)
type Dialer func() (net.Conn, error)

// A Reusable conn is a syncronized structure around a
// Dialer / net.Conn pair.  All net.Conn calls are wrapped
// around the underlying structure.  Errors are bubbled
// up, and trigger closure of the underlying socket (to
// be reopened on the next call)
type ReusableConn struct {
	lock          *sync.Mutex
	dialer        Dialer
	conn          net.Conn
	readDeadline  time.Time
	writeDeadline time.Time
}

// Create a new reusable connection with a sepcific dialer.
func NewReusableConnection(d Dialer) (c *ReusableConn) {
	return &ReusableConn{
		dialer: d,
		conn:   nil,
		lock:   &sync.Mutex{},
	}
}

// Dial is idempotent, and safe to call;
func (self *ReusableConn) Dial() (err error) {
	// log.Printf("Public Dial() called")
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.dial()
}

// Dial will redial if conn is nil, and set
// deadlines if they've been set by the caller.
// 
// It simply returns nil if the socket appears already connected
func (self *ReusableConn) dial() (err error) {
	// log.Printf("Private dial() called (%v)", self.conn)
	if self.conn == nil {
		self.conn, err = self.dialer()
		if err == nil && !self.readDeadline.IsZero() {
			err = self.setReadDeadline(self.readDeadline)
		}
		if err == nil && !self.writeDeadline.IsZero() {
			err = self.setWriteDeadline(self.writeDeadline)
		}
	}
	// log.Printf("Private dial() complete (%v)", self.conn)
	return
}

func (self *ReusableConn) close() (err error) {
	if self.conn != nil {
		err = self.conn.Close()
		self.conn = nil
	}
	return
}

// Unlike close on a traditional socket, no error
// is raised if you close a closed (nil) connection.
func (self *ReusableConn) Close() (err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.close()
}

// TODO: What's an appropriate responsde when we're not connected?
// ATM, we return whatever the other side says, or the nil net.Addr.
func (self *ReusableConn) RemoteAddr() (a net.Addr) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.conn != nil {
		a = self.conn.RemoteAddr()
	}
	return
}

// See RemoteAddr for notes.
func (self *ReusableConn) LocalAddr() (a net.Addr) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.conn != nil {
		a = self.conn.RemoteAddr()
	}
	return
}

func (self *ReusableConn) read(in []byte) (n int, err error) {
	err = self.dial()
	if err == nil {
		n, err = self.conn.Read(in)
		if err != nil {
			self.close()
		}
	}
	return
}

func (self *ReusableConn) write(in []byte) (n int, err error) {
	err = self.dial()
	if err == nil {
		n, err = self.conn.Write(in)
		if err != nil {
			self.close()
		}
	}
	return
}

// Read from the underlying connection, triggering a dial if needed.
// NB: For the expected case (HTTP), this shouldn't happen before the
// first Write.
func (self *ReusableConn) Read(in []byte) (n int, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.read(in)
}

// Write to the underlying connection, triggering a dial if needed.
func (self *ReusableConn) Write(out []byte) (n int, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.write(out)
}

func (self *ReusableConn) setReadDeadline(t time.Time) (err error) {
	err = self.dial()
	if err == nil {
		err = self.conn.SetReadDeadline(t)
		if err == nil {
			self.readDeadline = t
		}
	}
	return
}

func (self *ReusableConn) setWriteDeadline(t time.Time) (err error) {
	err = self.dial()
	if err == nil {
		err = self.conn.SetWriteDeadline(t)
		if err == nil {
			self.writeDeadline = t
		}
	}
	return
}

// Sets the read deadline on the underlying socket, as well
// as an internal flag for any future re-opened connections.
func (self *ReusableConn) SetReadDeadline(t time.Time) (err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.setReadDeadline(t)
}

// Sets the write deadline on the underlying socket, as well
// as an internal flag for any future re-opened connections.
func (self *ReusableConn) SetWriteDeadline(t time.Time) (err error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.setWriteDeadline(t)
}

// Convenience function for Set(Read|Write)Deadline
func (self *ReusableConn) SetDeadline(t time.Time) (err error) {
	err = self.SetReadDeadline(t)
	if err == nil {
		err = self.SetWriteDeadline(t)
	}
	return
}
