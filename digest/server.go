package digest

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

type HTTP2Server struct {
	sync.RWMutex
	*util.FunctionDaemon
	*logging.Logging
	bind             string
	host             string
	srv              *http.Server
	idleTimeout      time.Duration
	activeTimeout    time.Duration
	keepAliveTimeout time.Duration
	router           *mux.Router
}

func NewHTTP2Server(bind, host string, certs []tls.Certificate) (*HTTP2Server, error) {
	var getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)

	if len(certs) < 1 {
		if host == "localhost" || strings.HasPrefix(host, "127.0.0.") {
			if priv, err := util.GenerateED25519Privatekey(); err != nil {
				return nil, err
			} else if ct, err := util.GenerateTLSCerts("localhost", priv); err != nil {
				return nil, err
			} else {
				certs = ct
			}
		}
	}

	idleTimeout := time.Second * 10
	sv := &HTTP2Server{
		Logging: logging.NewLogging(func(c logging.Context) logging.Emitter {
			return c.Str("module", "http2-server")
		}),
		bind:             bind,
		host:             host,
		idleTimeout:      idleTimeout,     // TODO can be configurable
		activeTimeout:    time.Minute * 1, // TODO can be configurable
		keepAliveTimeout: time.Minute * 1, // TODO can be configurable
		router:           mux.NewRouter(),
	}

	if len(certs) > 0 {
		getCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return &certs[0], nil
		}
	} else {
		am := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(host),
		}
		getCertificate = am.GetCertificate
	}

	if srv, err := newHTTP2Server(sv, getCertificate); err != nil {
		return nil, err
	} else {
		sv.srv = srv
	}

	sv.FunctionDaemon = util.NewFunctionDaemon(sv.start, false)

	return sv, nil
}

func newHTTP2Server(
	sv *HTTP2Server,
	getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error),
) (*http.Server, error) {
	srv := &http.Server{
		Addr:         sv.bind,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Minute * 1,
		IdleTimeout:  sv.idleTimeout,
		TLSConfig: &tls.Config{
			GetCertificate: getCertificate,
			MinVersion:     tls.VersionTLS13,
		},
		// ErrorLog:  // TODO connect with http loggign
		Handler: network.HTTPLogHandler(sv.router, sv.Log()),
	}
	if err := http2.ConfigureServer(srv, &http2.Server{
		NewWriteScheduler: func() http2.WriteScheduler {
			return http2.NewPriorityWriteScheduler(nil)
		},
	}); err != nil {
		return nil, err
	} else {
		return srv, nil
	}
}

func (sv *HTTP2Server) Initialize() error {
	if ln, err := net.Listen("tcp", sv.bind); err != nil {
		return err
	} else if err := ln.Close(); err != nil {
		return err
	}

	root := sv.router.Name("root")
	root.Path("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		},
	)

	return nil
}

func (sv *HTTP2Server) SetHandler(handler http.Handler) {
	sv.srv.Handler = handler
}

func (sv *HTTP2Server) start(stopchan chan struct{}) error {
	if ln, err := net.Listen("tcp", sv.bind); err != nil {
		return err
	} else {
		errchan := make(chan error)
		sv.srv.ConnState = sv.idleTimeoutHook()
		go func() {
			errchan <- sv.srv.Serve(
				tls.NewListener(
					tcpKeepAliveListener{
						TCPListener:      ln.(*net.TCPListener),
						keepAliveTimeout: sv.keepAliveTimeout,
					},
					sv.srv.TLSConfig,
				),
			)
		}()

		select {
		case err := <-errchan:
			if err == nil || err == http.ErrServerClosed {
				sv.Log().Debug().Msg("server closed")

				return nil
			}

			sv.Log().Error().Err(err).Msg("something wrong")

			return err
		case <-stopchan:
			return sv.srv.Shutdown(context.Background())
		default:
			return nil
		}
	}
}

func (sv *HTTP2Server) idleTimeoutHook() func(net.Conn, http.ConnState) {
	var mu sync.Mutex
	m := map[net.Conn]*time.Timer{}
	return func(c net.Conn, cs http.ConnState) {
		mu.Lock()
		defer mu.Unlock()
		if t, ok := m[c]; ok {
			delete(m, c)
			t.Stop()
		}
		var d time.Duration
		switch cs {
		case http.StateNew, http.StateIdle:
			d = sv.idleTimeout
		case http.StateActive:
			d = sv.activeTimeout
		default:
			return
		}
		m[c] = time.AfterFunc(d, func() {
			sv.Log().Debug().Dur("idle-timeout", d).Str("remote", c.RemoteAddr().String()).Msg("closing idle conn after timeout")

			go func() {
				if err := c.Close(); err != nil {
					sv.Log().Error().Err(err).Msg("failed to close")
				}
			}()
		})
	}
}

type tcpKeepAliveListener struct {
	*net.TCPListener
	keepAliveTimeout time.Duration
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	if tc, err := ln.AcceptTCP(); err != nil {
		return nil, err
	} else {
		if err := tc.SetKeepAlive(true); err != nil {
			return nil, err
		}

		if err := tc.SetKeepAlivePeriod(ln.keepAliveTimeout); err != nil {
			return nil, err
		}

		return tc, nil
	}
}