/*
MIT License

Copyright (c) 2026 gounix

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package frontend

import (
        "fmt"
	"log/slog"
        "net/http"
        "registry-exporter/data"
        "registry-exporter/environ"
)

const (
	metric_prefix = "registry"
	promHeader = `# HELP registry_repo size of repo in bytes
# TYPE registry_repo gauge
# HELP registry_tag size of tag in bytes
# TYPE registry_tag gauge
# HELP registry_size size of complete registry
# TYPE registry_size gauge
`
)

func logRequest(r *http.Request) {
        slog.Info("frontend.logRequest", "Host", r.Host, "Method", r.Method, "Url", r.URL.Path, "UserAgent", r.UserAgent())
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {

        logRequest(r)
	fmt.Fprintf(w, promHeader)

	data.Registry.Mu.Lock()
        defer data.Registry.Mu.Unlock()

	slog.Info("frontend.metricsHandler start")
	for _, repo := range data.Registry.Repo {
		slog.Info("frontend.metricsHandler", "repo", repo.Name)

		str := fmt.Sprintf("%s_repo{registry=\"%s\", repo=\"%s\"} %d\n", metric_prefix, environ.Env.Registry, repo.Name, repo.Size)
		fmt.Fprintf(w, str)
		slog.Info("frontend.metricsHandler", "str", str)

		for _, tag := range repo.Tags {

			slog.Info("frontend.metricsHandler", "tag", tag.Tag)
			manifestListMetrics(w, repo, tag)
		}
	}
	str := fmt.Sprintf("%s_size{registry=\"%s\"} %d\n", metric_prefix, environ.Env.Registry, data.Registry.Size)
	fmt.Fprintf(w, str)
	slog.Info("frontend.metricsHandler", "str", str)
}

func manifestListMetrics(w http.ResponseWriter, repo data.RepositoryT, tag data.TagT) {
	for _, entry := range tag.ManifestList.Manifest {
		str := fmt.Sprintf("%s_tag{registry=\"%s\", repo=\"%s\", tag=\"%s\", arch=\"%s\", os=\"%s\", mediaType=\"%s\"} %d\n", 
			metric_prefix, environ.Env.Registry, repo.Name, tag.Tag, entry.Platform.Architecture, 
			entry.Platform.Os, entry.MediaType, entry.Size)
		fmt.Fprintf(w, str)
		slog.Info("frontend.manifestListMetrics", "str", str)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
        logRequest(r)
        if data.Alive() {
                fmt.Fprintf(w, "OK")
        } else {
                w.WriteHeader(http.StatusNotFound)
        }
}

func Server() {
        http.HandleFunc("/metrics", metricsHandler)
        http.HandleFunc("/health", healthHandler)

        addr := fmt.Sprintf(":%d", environ.Env.PortNumber)

	slog.Info("frontend.Server", "listen", addr)
        http.ListenAndServe(addr, nil)
}

