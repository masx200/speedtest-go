package web

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/masx200/speedtest-go/config"
	"github.com/masx200/speedtest-go/results"
	"github.com/pires/go-proxyproto"
	"github.com/quic-go/quic-go/http3"
	log "github.com/sirupsen/logrus"
)

const (
	// chunk size is 1 mib
	chunkSize = 1048576
)

//go:embed assets
var defaultAssets embed.FS

var (
	// generate random data for download test on start to minimize runtime overhead
	randomData = getRandomData(chunkSize)
)

func ListenAndServe(conf *config.Config) error {
	r := chi.NewRouter()
	var httpsPort = conf.Port
	if conf.EnableHTTP3 && conf.EnableTLS {
		r.Use(func(h http.Handler) http.Handler {
			fn := func(writer http.ResponseWriter, r *http.Request) {
				writer.Header().Add("Alt-Svc",
					"h3=\":"+fmt.Sprint(httpsPort)+"\";ma=86400,h3-29=\":"+fmt.Sprint(httpsPort)+"\";ma=86400,h3-27=\":"+fmt.Sprint(httpsPort)+"\";ma=86400",
				)
				// Delete any ETag headers that may have been set
				// for _, v := range etagHeaders {
				// 	if r.Header.Get(v) != "" {
				// 		r.Header.Del(v)
				// 	}
				// }

				// // Set our NoCache headers
				// for k, v := range noCacheHeaders {
				// 	w.Header().Set(k, v)
				// }

				h.ServeHTTP(writer, r)
			}
			// return fn
			return http.HandlerFunc(fn)
			// c.Writer.Header().Add("Alt-Svc",
			// 	"h3=\":"+fmt.Sprint(httpsPort)+"\";ma=86400,h3-29=\":"+fmt.Sprint(httpsPort)+"\";ma=86400,h3-27=\":"+fmt.Sprint(httpsPort)+"\";ma=86400",
			// )
			// c.Next()
		},
			func(h http.Handler) http.Handler {
				fn := func(writer http.ResponseWriter, r *http.Request) {
					/* ctx.W */ writer.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
					// ctx.Next()
					h.ServeHTTP(writer, r)
				}
				return http.HandlerFunc(fn)
			})

	}

	r.Use(middleware.RealIP)
	r.Use(middleware.GetHead)

	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{"*"},
	})

	r.Use(cs.Handler)
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)

	var assetFS http.FileSystem
	if fi, err := os.Stat(conf.AssetsPath); os.IsNotExist(err) || !fi.IsDir() {
		log.Warnf("Configured asset path %s does not exist or is not a directory, using default assets", conf.AssetsPath)
		sub, err := fs.Sub(defaultAssets, "assets")
		if err != nil {
			log.Fatalf("Failed when processing default assets: %s", err)
		}
		assetFS = http.FS(sub)
	} else {
		assetFS = justFilesFilesystem{fs: http.Dir(conf.AssetsPath), readDirBatchSize: 2}
	}

	r.Get("/*", pages(assetFS))
	r.HandleFunc("/empty", empty)
	r.HandleFunc("/backend/empty", empty)
	r.Get("/garbage", garbage)
	r.Get("/backend/garbage", garbage)
	r.Get("/getIP", getIP)
	r.Get("/backend/getIP", getIP)
	r.Get("/results", results.DrawPNG)
	r.Get("/results/", results.DrawPNG)
	r.Get("/backend/results", results.DrawPNG)
	r.Get("/backend/results/", results.DrawPNG)
	r.Post("/results/telemetry", results.Record)
	r.Post("/backend/results/telemetry", results.Record)
	r.HandleFunc("/stats", results.Stats)
	r.HandleFunc("/backend/stats", results.Stats)

	// PHP frontend default values compatibility
	r.HandleFunc("/empty.php", empty)
	r.HandleFunc("/backend/empty.php", empty)
	r.Get("/garbage.php", garbage)
	r.Get("/backend/garbage.php", garbage)
	r.Get("/getIP.php", getIP)
	r.Get("/backend/getIP.php", getIP)
	r.Post("/results/telemetry.php", results.Record)
	r.Post("/backend/results/telemetry.php", results.Record)
	r.HandleFunc("/stats.php", results.Stats)
	r.HandleFunc("/backend/stats.php", results.Stats)

	go listenProxyProtocol(conf, r)

	var s error

	addr := net.JoinHostPort(conf.BindAddress, conf.Port)
	log.Infof("Starting backend server on %s", addr)

	// TLS and HTTP/2.
	if conf.EnableTLS {
		log.Info("Use TLS connection.")

		if conf.EnableHTTP3 {
			return http3.ListenAndServe(addr, conf.TLSCertFile, conf.TLSKeyFile, r)
		}
		if !(conf.EnableHTTP2) {
			srv := &http.Server{
				Addr:         addr,
				Handler:      r,
				TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
			}
			s = srv.ListenAndServeTLS(conf.TLSCertFile, conf.TLSKeyFile)
		} else {
			s = http.ListenAndServeTLS(addr, conf.TLSCertFile, conf.TLSKeyFile, r)
		}
	} else {
		if conf.EnableHTTP2 {
			log.Errorf("TLS is mandatory for HTTP/2. Ignore settings that enable HTTP/2.")
		}
		s = http.ListenAndServe(addr, r)
	}

	return s
}

func listenProxyProtocol(conf *config.Config, r *chi.Mux) {
	if conf.ProxyProtocolPort != "0" {
		addr := net.JoinHostPort(conf.BindAddress, conf.ProxyProtocolPort)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Cannot listen on proxy protocol port %s: %s", conf.ProxyProtocolPort, err)
		}

		pl := &proxyproto.Listener{Listener: l}
		defer pl.Close()

		log.Infof("Starting proxy protocol listener on %s", addr)
		log.Fatal(http.Serve(pl, r))
	}
}

func pages(fs http.FileSystem) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			r.RequestURI = "/index.html"
		}

		http.FileServer(fs).ServeHTTP(w, r)
	}

	return fn
}

func empty(w http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

func garbage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=random.dat")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// chunk size set to 4 by default
	chunks := 4

	ckSize := r.FormValue("ckSize")
	if ckSize != "" {
		i, err := strconv.ParseInt(ckSize, 10, 64)
		if err != nil {
			log.Errorf("Invalid chunk size: %s", ckSize)
			log.Warnf("Will use default value %d", chunks)
		} else {
			// limit max chunk size to 1024
			if i > 1024 {
				chunks = 1024
			} else {
				chunks = int(i)
			}
		}
	}

	for i := 0; i < chunks; i++ {
		if _, err := w.Write(randomData); err != nil {
			log.Errorf("Error writing back to client at chunk number %d: %s", i, err)
			break
		}
	}
}

func getIP(w http.ResponseWriter, r *http.Request) {
	var ret results.Result

	clientIP := r.RemoteAddr
	clientIP = strings.ReplaceAll(clientIP, "::ffff:", "")

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		clientIP = ip
	}

	isSpecialIP := true
	switch {
	case clientIP == "::1":
		ret.ProcessedString = clientIP + " - localhost IPv6 access"
	case strings.HasPrefix(clientIP, "fe80:"):
		ret.ProcessedString = clientIP + " - link-local IPv6 access"
	case strings.HasPrefix(clientIP, "127."):
		ret.ProcessedString = clientIP + " - localhost IPv4 access"
	case strings.HasPrefix(clientIP, "10."):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "192.168"):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "169.254"):
		ret.ProcessedString = clientIP + " - link-local IPv4 access"
	case regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - CGNAT IPv4 access"
	default:
		isSpecialIP = false
	}

	if isSpecialIP {
		b, _ := json.Marshal(&ret)
		if _, err := w.Write(b); err != nil {
			log.Errorf("Error writing to client: %s", err)
		}
		return
	}

	getISPInfo := r.FormValue("isp") == "true"
	distanceUnit := r.FormValue("distance")

	ret.ProcessedString = clientIP

	if getISPInfo {
		ispInfo := getIPInfo(clientIP)
		ret.RawISPInfo = ispInfo

		removeRegexp := regexp.MustCompile(`AS\d+\s`)
		isp := removeRegexp.ReplaceAllString(ispInfo.Organization, "")

		if isp == "" {
			isp = "Unknown ISP"
		}

		if ispInfo.Country != "" {
			isp += ", " + ispInfo.Country
		}

		if ispInfo.Location != "" {
			isp += " (" + calculateDistance(ispInfo.Location, distanceUnit) + ")"
		}

		ret.ProcessedString += " - " + isp
	}

	render.JSON(w, r, ret)
}
