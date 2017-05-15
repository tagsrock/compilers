package perform

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/monax/compilers/definitions"

	"path/filepath"

	"crypto/tls"
	"io"
	"net"

	"github.com/monax/cli/config"
	"github.com/monax/cli/log"
)

var BinariesPath = filepath.Join(config.MonaxRoot, "binaries")

// Start the compile server. Takes either or both of addrInsecure or addrSecure
// to run on HTTP or HTTPS respectively. If addrSecure is passed a certFile and
// keyFile path must be passed for TLS support.
//
// Returns an io.Closer that can be used to close the underlying http(s)
// listeners and a shutdown channel over which a value is sent if the server is
// shutdown. That value will be
func StartServer(addrInsecure, addrSecure, certFile, keyFile string) (io.Closer,
	chan error) {
	log.Warn("Hello I'm the marmots' compilers server")
	err := config.InitMonaxDir()
	if err != nil {
		log.Errorf("Error making Monax CLI directories: %s", err)
		os.Exit(1)
	}
	err = config.InitDataDir(BinariesPath)
	if err != nil {
		log.Errorf("Error making Monax Keys directories: %s", err)
		os.Exit(1)
	}

	// Routes on dedicated mux
	mux := http.NewServeMux()
	mux.HandleFunc("/", CompileHandler)
	mux.HandleFunc("/binaries", BinaryHandler)
	srv := &http.Server{Handler: mux}

	var listeners netListeners

	// Use SSL ?
	log.Debug(certFile)

	// Returns any error from listeners, give it buffer the same size as the
	// number of listeners to so listener goroutines don't block
	shutdownChan := make(chan error, 2)
	if addrSecure != "" {
		log.Debug("Using HTTPS")
		log.WithField("=>", addrSecure).Debug("Listening on...")

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Errorf("Could not load TLS certificate: %s", err)
			os.Exit(1)
		}
		httpsListener, err := tls.Listen("tcp", addrSecure,
			&tls.Config{Certificates: []tls.Certificate{cert}})
		if err != nil {
			log.Errorf("Could not create HTTPS listener: %s", err)
			os.Exit(1)
		}
		listeners = append(listeners, httpsListener)
		go func() {
			shutdownChan <- srv.Serve(httpsListener)
		}()
	}
	if addrInsecure != "" {
		log.Debug("Using HTTP")
		log.WithField("=>", addrInsecure).Debug("Listening on...")
		httpListener, err := net.Listen("tcp", addrInsecure)
		if err != nil {
			log.Errorf("Could not create HTTP listener: %s", err)
			os.Exit(1)
		}
		listeners = append(listeners, httpListener)
		go func() {
			shutdownChan <- srv.Serve(httpListener)
		}()
	}
	return listeners, shutdownChan
}

// Main http request handler
// Read request, compile, build response object, write
func CompileHandler(w http.ResponseWriter, r *http.Request) {
	resp := compileResponse(w, r)
	if resp == nil {
		return
	}
	respJ, err := json.Marshal(resp)
	if err != nil {
		log.Errorln("failed to marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(respJ)
}

func BinaryHandler(w http.ResponseWriter, r *http.Request) {
	// read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorln("err on read http request body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// unmarshall body into req struct
	req := new(definitions.BinaryRequest)
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Errorln("err on json unmarshal of request", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	resp := linkBinaries(req)
	respJ, err := json.Marshal(resp)
	if err != nil {
		log.Errorln("failed to marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(respJ)
}

// read in the files from the request, compile them
func compileResponse(w http.ResponseWriter, r *http.Request) *Response {
	// read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorln("err on read http request body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	// unmarshall body into req struct
	req := new(definitions.Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Errorln("err on json unmarshal of request", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	log.WithFields(log.Fields{
		"lang": req.Language,
		// "script": string(req.Script),
		"libs": req.Libraries,
		"incl": req.Includes,
	}).Debug("New Request")

	cached := CheckCached(req.Includes, req.Language)

	log.WithField("cached?", cached).Debug("Cached Item(s)")

	var resp *Response
	// if everything is cached, no need for request
	if cached {
		resp, err = CachedResponse(req.Includes, req.Language)
		if err != nil {
			log.Errorln("err during caching response", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
	} else {
		resp = compile(req)
		resp.CacheNewResponse(*req)
	}

	return resp
}

type netListeners []net.Listener

var _ io.Closer = netListeners(nil)

func (listeners netListeners) Close() error {
	for _, listener := range listeners {
		err := listener.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
