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
	"encoding/json"
	"github.com/gounix/goregistry/v2"
	"log/slog"
	"registry-exporter/data"
	"registry-exporter/environ"
	"time"
)

type (
        ConfigBlobT struct {
                Created time.Time `json:"created"`
                Arch    string    `json:"architecture"`
                Os      string    `json:"os"`
        }

)

func setupRegistry() goregistry.RegistryT{
        var reg goregistry.RegistryT

        reg.Scheme = environ.Env.Scheme
        reg.TlsVerify = true
        reg.Host = environ.Env.Registry
        reg.Image = ""
        reg.Regcred.User = ""
        reg.Regcred.Passwd = ""

        return reg
}

func fetchManifestSize(manifest goregistry.ManifestT) int64 {
	size := int64(0)
	for _, layer := range manifest.Json.Layers {
		size = size + layer.Size
		addDigest(layer.Size, layer.Digest)
	}
	return size
}

func fetchSize(reg goregistry.RegistryT, repo string, tag string) int64 {
	list, err := reg.GetManifestList(tag)
	if err != nil {
		// see if it is a manifest
		manifest, err := reg.GetManifest("application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json", tag)
		if err != nil {
			slog.Error("registry.fetchSize", "reg.GetManifest", err)
			return 0
		}
		return fetchManifestSize(manifest)
	}
	size := int64(0)
	for _, entry := range list.Json.Manifest {
		if entry.Platform.Architecture != "unknown" {
			manifest, err := reg.GetManifest(entry.MediaType, entry.Digest)
			if err != nil {
				slog.Error("registry.fetchSize", "reg.GetManifest", err)
				return 0
			}
			size = size + fetchManifestSize(manifest)
		}
	}
	return size
}

func fetchConfigBlob(reg goregistry.RegistryT, accept string, digest string) (ConfigBlobT, error) {
	var dat ConfigBlobT

	blob, err := reg.GetBlob(accept, digest)
	err = json.Unmarshal(blob.Raw, &dat)
	if err != nil {
		slog.Error("registry.fetchConfigBlob", "err", err)
		return dat, err
	}
        slog.Info("registry.fetchConfigBlob", "architecture", dat.Arch, "os", dat.Os)
	return dat, nil
}

func fetchManifest(reg goregistry.RegistryT, accept string, tag string) (goregistry.ArchManifestT, error) {
	var am goregistry.ArchManifestT
	slog.Info("registry.fetchManifest", "tag", tag)
	manif, err := reg.GetManifest(accept, tag)
	if err != nil {
		slog.Error("registry.fetchManifest", "err", err)
		return am, err
	}

	config, err := fetchConfigBlob(reg, manif.Json.Config.MediaType, manif.Json.Config.Digest)
	if err != nil {
		slog.Error("registry.fetchManifest", "err", err)
	}
	am.Platform.Architecture = config.Arch
	am.Platform.Os = config.Os
	am.Size = fetchManifestSize(manif)
	am.MediaType = manif.Json.Config.MediaType
	slog.Info("registry.fetchManifest", "image", reg.Image, "tag", tag, "arch", am.Platform.Architecture, "size", am.Size)
	return am, nil
}

func fetchManifestList(reg goregistry.RegistryT, repo string, tag string) (data.ManifestListT, error)  {
	var ml data.ManifestListT
	slog.Info("registry.fetchManifestList", "repo", repo, "tag", tag)
	list, err := reg.GetManifestList(tag)
	if err != nil {
		var am goregistry.ArchManifestT
		// see if it is a manifest
		am, err := fetchManifest(reg, "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json", tag)
		if err != nil {
			slog.Error("registry.fetchManifestList single manifest", "err", err)
			return ml, err
		}
		ml.Manifest = append(ml.Manifest, am)
		slog.Info("registry.fetchManifestList", "image", reg.Image, "tag", tag, "arch", am.Platform.Architecture, "size", am.Size)
		return ml, nil

	}
	for _, entry := range list.Json.Manifest {
		new, _ := fetchManifest(reg, entry.MediaType, entry.Digest)
		ml.Manifest = append(ml.Manifest, new)
		slog.Info("registry.fetchManifestList", "image", reg.Image, "tag", tag, "arch", new.Platform.Architecture, "size", new.Size)
	}
	return ml, nil
}

func fetchTags(reg goregistry.RegistryT, repo string) []data.TagT {
	var retList []data.TagT

	err := reg.AcquireToken()
	if err != nil {
                slog.Error("registry.fetchTags AcquireToken", "err", err)
                return []data.TagT{}
        }
	reg.Image = repo
	newProject()
        tagList,_ := reg.GetVersions(".*", false)
        for _, tag := range tagList {
		var retEntry data.TagT

                slog.Info("registry.fetchTags", "repo", repo, "tag", tag)
		retEntry.Tag = tag
		retEntry.Size = fetchSize(reg, repo, tag)
		list, _ := fetchManifestList(reg, repo, tag)
		retEntry.ManifestList = list
		retList = append(retList, retEntry)
        }
        slog.Info("registry.fetchTags", "entries", len(retList))
	return retList
}

func fetchCatalog(reg goregistry.RegistryT) {

	err := reg.AcquireCatalogToken()
	if err != nil {
                slog.Error("registry.fetchCatalog AcquireCatalogToken", "err", err)
                return
        }
        list, err := reg.GetCatalog(".*", false)
        if err != nil {
                slog.Error("registry.fetchCatalog", "err", err)
        }
        slog.Info("registry.fetchCatalog", "matched", list)

	newTotal()
	repoList := []data.RepositoryT{}
	for _, repository := range list {
		var repo data.RepositoryT
		repo.Name = repository
		repo.Tags = fetchTags(reg, repository)
		repo.Size = getProjectSize()
		repoList = append(repoList, repo)
	}
        slog.Info("registry.fetchCatalog", "stored", len(repoList))

	data.Registry.Mu.Lock()
	defer data.Registry.Mu.Unlock()
	data.Registry.Size = getTotalSize()
	data.Registry.Repo = repoList
	data.Registry.Timestamp = time.Now()
}

func Fetch() {
	reg := setupRegistry()
	// to keep the probes happy, the refresh loop can take some time
	data.Registry.Timestamp = time.Now()
        for {
		fetchCatalog(reg)
		slog.Info("registry.Fetch", "sleep", environ.Env.RefreshSeconds)
		time.Sleep(time.Duration(environ.Env.RefreshSeconds) * time.Second)
        }
}

