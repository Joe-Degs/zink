package zinc

import "net"

type ZinkError = requestWrapper

func (e ZinkError) Error() string {
	return string(e.data)
}

func NewError(msg string) *ZinkError {
	return &ZinkError{typ: Error, data: []byte(msg)}
}

var (
	UnImplementedEndPoint = NewError("unimplemented endpoint")
	UnknownPacketType     = NewError("unknown packet type")
)

func ErrrorWithAddr(err *ZinkError, addr *net.UDPAddr) *ZinkError {
	err.addr = addr
	return err
}
