package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	socks5Version = byte(0x05)

	authNotRequired = byte(0x00)

	cmdTCPConnect = byte(0x01)

	statusSuccess       = byte(0x00)
	statusCmdNotSupport = byte(0x07)
)

func Socks5Auth(client net.Conn) (err error) {
	buf := make([]byte, 256)

	if _, err = io.ReadFull(client, buf[:2]); err != nil {
		return fmt.Errorf("read auth header error: %v", err)
	}

	ver, nMethods := buf[0], buf[1]
	if ver != socks5Version {
		return fmt.Errorf("invalid version: 0x%02x", ver)
	}

	if _, err = io.ReadFull(client, buf[:nMethods]); err != nil {
		return fmt.Errorf("read methods error: %v", err)
	}

	if _, err = client.Write([]byte{socks5Version, authNotRequired}); err != nil {
		return fmt.Errorf("write auth response error: %v", err)
	}

	return nil
}

func writeConnectResponse(client net.Conn, status byte) (int, error) {
	return client.Write([]byte{socks5Version, status, byte(0x00), byte(0x01),
		byte(0x00), byte(0x00), byte(0x00), byte(0x00), byte(0x00), byte(0x00)})
}

type Dest struct {
	addr string
	port uint16
}

func (d Dest) URL() string {
	return d.Scheme() + "://" + d.String()
}

func (d Dest) Scheme() string {
	if d.port == 443 {
		return "https"
	}
	return "http"
}

func (d Dest) String() string {
	return fmt.Sprintf("%s:%d", d.addr, d.port)
}

func Socks5Connect(client net.Conn) (dest Dest, err error) {
	buf := make([]byte, 256)

	if _, err = io.ReadFull(client, buf[:4]); err != nil {
		writeConnectResponse(client, statusCmdNotSupport)
		err = fmt.Errorf("read connect header error: %v", err)
		return
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != socks5Version {
		writeConnectResponse(client, statusCmdNotSupport)
		err = fmt.Errorf("invalid version: 0x%02x", ver)
		return
	}

	if cmd != cmdTCPConnect {
		writeConnectResponse(client, statusCmdNotSupport)
		err = fmt.Errorf("invalid command: 0x%02x", ver)
		return
	}

	addr := ""
	switch atyp {
	case 0x01:
		_, err = io.ReadFull(client, buf[:4])
		if err != nil {
			writeConnectResponse(client, statusCmdNotSupport)
			err = fmt.Errorf("invalid IPv4: %v", err)
			return
		}
		addr = net.IP(buf[:4]).String()

	case 0x03:
		_, err = io.ReadFull(client, buf[:1])
		if err != nil {
			writeConnectResponse(client, statusCmdNotSupport)
			err = fmt.Errorf("invalid hostname: %v", err)
			return
		}
		addrLen := buf[0]

		_, err = io.ReadFull(client, buf[:addrLen])
		if err != nil {
			writeConnectResponse(client, statusCmdNotSupport)
			err = fmt.Errorf("invalid hostname: %v", err)
			return
		}
		addr = string(buf[:addrLen])

	case 0x04:
		_, err = io.ReadFull(client, buf[:16])
		if err != nil {
			writeConnectResponse(client, statusCmdNotSupport)
			err = fmt.Errorf("invalid IPv6: %v", err)
			return
		}
		addr = net.IP(buf[:16]).String()

	default:
		writeConnectResponse(client, statusCmdNotSupport)
		err = fmt.Errorf("invalid address type: 0x%02x", atyp)
		return
	}

	if _, err = io.ReadFull(client, buf[:2]); err != nil {
		writeConnectResponse(client, statusCmdNotSupport)
		err = fmt.Errorf("invalid port: %v", err)
		return
	}

	port := binary.BigEndian.Uint16(buf[:2])

	dest = Dest{
		addr: addr,
		port: port,
	}

	_, err = writeConnectResponse(client, statusSuccess)
	if err != nil {
		err = fmt.Errorf("write response: %v", err)
	}

	return
}

func forward(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	io.Copy(dst, src)
}

func Socks5Forward(client, target net.Conn) {
	go forward(target, client)
	go forward(client, target)
}
