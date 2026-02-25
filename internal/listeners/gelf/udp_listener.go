package gelf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

func StartUdpListenerSupervisor(ctx context.Context, port int, wg *sync.WaitGroup) {
	defer wg.Done()

	packetsChan := make(chan Packet, 1024)

	var workersWaitGroup sync.WaitGroup

	// Start UdpListener workers
	for i := 0; i < 5; i++ {
		workersWaitGroup.Add(1)
		go StartWorkerSupervisor(ctx, i, packetsChan, &workersWaitGroup)
	}

	for {
		slog.InfoContext(ctx, "Starting UdpListener", "port", port)

		err := runUdpListener(ctx, port, packetsChan)

		if err == nil {
			break
		}

		slog.WarnContext(ctx, "UdpListener failed. Restarting in 5s ...", "port", port, "err", err)
		time.Sleep(5 * time.Second)
	}

	close(packetsChan)      // Tell listener workers no more data is coming
	workersWaitGroup.Wait() // Wait for workers to finish the queue
}

func runUdpListener(ctx context.Context, port int, packets chan<- Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Recovered UdpListener from panic", "port", port, "recover", r)
			err = errors.New(r.(string))
		}
	}()

	// Resolve the UDP address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.ErrorContext(ctx, "Resolve error", "err", err)
		return err
	}

	conn, err := net.ListenPacket("udp", addr.String())
	if err != nil {
		slog.ErrorContext(ctx, "Listen error", "err", err)
		return err
	}

	defer func(conn net.PacketConn) {
		err := conn.Close()
		if err != nil {
			slog.ErrorContext(ctx, "Close error 1", "err", err)
		}
	}(conn)

	slog.InfoContext(ctx, "UdpListener started", "port", port)

	buffer := make([]byte, 8192)
	// Main Listener Loop
	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Stopping UdpListener...")
			return nil
		default:
			if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
				slog.ErrorContext(ctx, "SetReadDeadline error", "err", err)
				continue
			}

			n, remoteAddr, err := conn.ReadFrom(buffer)
			if err != nil {

				if !(err.Error() != "i/o timeout") {
					slog.ErrorContext(ctx, "Read error", "err", err)
				}

				continue
			}

			data := make([]byte, n)
			copy(data, buffer[:n])

			packets <- Packet{
				Addr: remoteAddr.String(),
				Data: data,
			}
		}
	}
}
