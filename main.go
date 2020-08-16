package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const indexHTML = `<!DOCTYPE html>
<script src="wasm_exec.js"></script><script>
(async () => {
  const resp = await fetch('main.wasm');
  if (!resp.ok) {
    const pre = document.createElement('pre');
    pre.innerText = await resp.text();
    document.body.appendChild(pre);
    return;
  }
  const src = await resp.arrayBuffer();
  const go = new Go();
  const result = await WebAssembly.instantiate(src, go.importObject);
  go.run(result.instance);
})();
</script>
`

var (
	flagHTTP        = flag.String("http", ":8080", "HTTP bind address to serve")
	flagTags        = flag.String("tags", "", "Build tags")
	flagAllowOrigin = flag.String("allow-origin", "", "Allow specified origin (or * for all origins) to make requests to this server")
	flagNoDebug 	= flag.Bool("no-debug", false, "Disable outputting debug symbols. Avoiding debug symbols can have a big impact on generated binary size, reducing them by more than half.")
)

var tmpOutputDir = ""

func ensureTmpOutputDir() (string, error) {
	if tmpOutputDir != "" {
		return tmpOutputDir, nil
	}

	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	tmpOutputDir = tmp
	return tmpOutputDir, nil
}


func handle(w http.ResponseWriter, r *http.Request) {
	if *flagAllowOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", *flagAllowOrigin)
	}

	output, err := ensureTmpOutputDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	upath := r.URL.Path[1:]
	fpath := filepath.Join(".", filepath.Base(upath))
	workdir := "."

	if !strings.HasSuffix(r.URL.Path, "/") {
		fi, err := os.Stat(fpath)
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fi != nil && fi.IsDir() {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
			return
		}
	}

	switch filepath.Base(fpath) {
	case "index.html", ".":
		if _, err := os.Stat(fpath); err != nil && !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader([]byte(indexHTML)))
		return
	case "wasm_exec.js":
		if _, err := os.Stat(fpath); err != nil && !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tinygoRootEnv := exec.Command("tinygo", "env", "TINYGOROOT")
		rootEnv, err := tinygoRootEnv.CombinedOutput()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		f := filepath.Join(strings.TrimSuffix(string(rootEnv), "\n"), "targets", "wasm_exec.js")
		http.ServeFile(w, r, f)
		return
	case "main.wasm":
		if _, err := os.Stat(fpath); err != nil && !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// tinygo build -o main.wasm -target wasm .
		args := []string{"build", "-o", filepath.Join(output, "main.wasm"), "-target", "wasm"}
		if *flagNoDebug {
			args = append(args, "-no-debug")
		}
		if *flagTags != "" {
			args = append(args, "-tags", *flagTags)
		}
		if len(flag.Args()) > 0 {
			args = append(args, flag.Args()[0])
		} else {
			args = append(args, ".")
		}
		log.Print("tinygo ", strings.Join(args, " "))
		cmdBuild := exec.Command("tinygo", args...)
		cmdBuild.Dir = workdir
		out, err := cmdBuild.CombinedOutput()
		if err != nil {
			log.Print(err)
			log.Print(string(out))
			http.Error(w, string(out), http.StatusInternalServerError)
			return
		}
		if len(out) > 0 {
			log.Print(string(out))
		}

		f, err := os.Open(filepath.Join(output, "main.wasm"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		http.ServeContent(w, r, "main.wasm", time.Now(), f)
		return
	}

	http.ServeFile(w, r, filepath.Join(".", r.URL.Path))
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(*flagHTTP, nil))
}