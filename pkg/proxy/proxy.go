package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/goproxyio/goproxy/pkg/modfetch"
	"github.com/goproxyio/goproxy/pkg/modfetch/codehost"
	"github.com/goproxyio/goproxy/pkg/module"
)

var cacheDir string
var innerHandle http.Handler

func NewProxy(cache string) http.Handler {
	modfetch.PkgMod = filepath.Join(cache, "pkg", "mod")
	codehost.WorkRoot = filepath.Join(modfetch.PkgMod, "cache", "vcs")

	cacheDir = filepath.Join(modfetch.PkgMod, "cache", "download")
	innerHandle = http.FileServer(http.Dir(cacheDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("goproxy: %s download %s\n", r.RemoteAddr, r.URL.Path)
		if _, err := os.Stat(filepath.Join(cacheDir, r.URL.Path)); err != nil {
			suffix := path.Ext(r.URL.Path)
			if suffix == ".info" || suffix == ".mod" || suffix == ".zip" {
				mod := strings.Split(r.URL.Path, "/@v/")
				if len(mod) != 2 {
					ReturnBadRequest(w, fmt.Errorf("bad module path:%s", r.URL.Path))
					return
				}
				version := strings.TrimSuffix(mod[1], suffix)
				version, err = module.DecodeVersion(version)
				if err != nil {
					ReturnServerError(w, err)
					return
				}
				modPath := strings.TrimPrefix(mod[0], "/")
				modPath, err := module.DecodePath(modPath)
				if err != nil {
					ReturnServerError(w, err)
					return
				}
				// ignore the error, incorrect tag may be given
				// forward to inner.ServeHTTP
				if err := downloadMod(w, r, modPath, version, suffix); err != nil {
					errLogger.Printf("download get err %s", err)
				}
			}

			// fetch latest version
			if strings.HasSuffix(r.URL.Path, "/@latest") {
				modPath := strings.TrimSuffix(r.URL.Path, "/@latest")
				modPath = strings.TrimPrefix(modPath, "/")
				modPath, err := module.DecodePath(modPath)
				if err != nil {
					ReturnServerError(w, err)
					return
				}
				repo, err := modfetch.Lookup(modPath)
				if err != nil {
					errLogger.Printf("lookup failed: %v", err)
					ReturnServerError(w, err)
					return
				}
				rev, err := repo.Stat("latest")
				if err != nil {
					errLogger.Printf("latest failed: %v", err)
					return
				}
				if err := downloadMod(w, r, modPath, rev.Version, ""); err != nil {
					errLogger.Printf("download get err %s", err)
				}
				err = json.NewEncoder(w).Encode(rev)
				if err != nil {
					errLogger.Printf("json encode failed %s", err)
				}
				return
			}

			if strings.HasSuffix(r.URL.Path, "/@v/list") {
				// TODO
				_, _ = w.Write([]byte(""))
				return
			}
		}
		innerHandle.ServeHTTP(w, r)
	})
}

type goModule struct {
	Path     string `json:"path"`     // module path
	Version  string `json:"version"`  // module version
	Error    string `json:"error"`    // error loading module
	Info     string `json:"info"`     // absolute path to cached .info file
	GoMod    string `json:"goMod"`    // absolute path to cached .mod file
	Zip      string `json:"zip"`      // absolute path to cached .zip file
	Dir      string `json:"dir"`      // absolute path to cached source root directory
	Sum      string `json:"sum"`      // checksum for path, version (as in go.sum)
	GoModSum string `json:"goModSum"` // checksum for go.mod (as in go.sum)
}

func downloadMod(w http.ResponseWriter, r *http.Request, path, version, suffix string) error {
	cmd := exec.Command("go", "mod", "download", "-json", path+"@"+version)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "goproxy: download %s stderr:\n%s", path, string(stderr.Bytes()))
		return err
	}

	var m goModule
	if err := json.NewDecoder(strings.NewReader(string(stdout.Bytes()))).Decode(&m); err != nil {
		return err
	}
	var p string
	if suffix == "" {
		return nil
	}
	if suffix == ".info" {
		p = m.Info
	}
	if suffix == ".mod" {
		p = m.GoMod
	}
	if suffix == ".zip" {
		p = m.Zip
	}
	p = strings.TrimPrefix(p, cacheDir)
	h := r.Host
	scheme := "http:"
	if r.TLS != nil {
		scheme = "https:"
	}
	url := fmt.Sprintf("%s//%s/%s", scheme, h, p)
	http.Redirect(w, r, url, 302)
	return nil
}
