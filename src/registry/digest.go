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

package registry

import (
        "log/slog"
)

type (
        layerT struct {
                size   int64
                digest string
        }
)

var project []layerT
var total   []layerT

func addToList(list []layerT, size int64, digest string) []layerT {
	var new layerT

	for _, entry := range list {
		if entry.digest == digest {
			slog.Info("registry.addToList", "found", digest)
			return list
		}
	}
	
	new.size = size
	new.digest = digest

	slog.Info("registry.addToList", "not found", digest)
	return append(list, new)
}

func addDigest(size int64, digest string) {
	project = addToList(project, size, digest)
	slog.Info("registry.addDigest", "len_project", len(project))
	total = addToList(total, size, digest)
	slog.Info("registry.addDigest", "len_total", len(project))
}

func newProject() {
	project = []layerT{}
}

func newTotal() {
	total = []layerT{}
}

func getListSize(list []layerT) int64 {
	var sum int64

	for _, entry := range list {
		sum = sum + entry.size
	}
	return sum
}

func getProjectSize() int64 {
	return getListSize(project)
}

func getTotalSize() int64 {
	return getListSize(total)
}

