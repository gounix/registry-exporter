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

package data

import (
	"github.com/gounix/goregistry/v2"
	"log/slog"
	"math"
	"sync"
	"time"
	"registry-exporter/environ"
)

type (
//	LayerT struct {
//		Size   int64
//		Digest string
//	}
//	ManifestT struct {
//		Arch   string
//		Layers []LayerT
//	}
	ManifestListT struct {
		Manifest []goregistry.ArchManifestT
	}
	TagT struct {
		Tag             string
		RawManifestList goregistry.ManifestListT // one of ManifestList or Manifest should be filled
		RawManifest     goregistry.ManifestT     // the other should be nil
		ManifestList    ManifestListT
		Size            int64         // computed
	}
	RepositoryT struct {
		Name string
		Tags []TagT
		Size int64      // computed
	}
	RegistryT struct {
		Mu          sync.Mutex
		Repo        []RepositoryT
		Size        int64         // computed
		Timestamp   time.Time
	}
)

var Registry RegistryT

func Alive() bool {
	Registry.Mu.Lock()
        defer Registry.Mu.Unlock()

        interval := environ.Env.RefreshSeconds
        now := time.Now()
        diff := now.Sub(Registry.Timestamp)

        // consider the producer dead after it missed 2 intervals
        isOK := diff.Seconds() < float64(2 * interval)
        slog.Info("data.Alive", "age(seconds)", math.Floor(diff.Seconds()), "OK", isOK)

        return isOK
}


