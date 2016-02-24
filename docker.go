package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/reference"
)

const (
	dockerPrefix       = "docker://"
	dockerHostname     = "docker.io"
	dockerRegistry     = "registry-1.docker.io"
	dockerAuthRegistry = "https://index.docker.io/v1/"
)

var validHex = regexp.MustCompile(`^([a-f0-9]{64})$`)

type dockerImage struct {
	ref      reference.Named
	tag      string
	registry string
	username string
	password string
}

func (i *dockerImage) Kind() Kind {
	return KindDocker
}

// will support v1 one day...
type manifest interface {
	GetLayers() []string
}

type manifestSchema1 struct {
	FSLayers []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
	History []struct {
		V1Compatibility string `json:"v1Compatibility"`
	} `json:"history"`
	// TODO(runcom) verify the downloaded manifest
	//Signature []byte `json:"signature"`
}

func (m *manifestSchema1) GetLayers() []string {
	layers := make([]string, len(m.FSLayers))
	for i, layer := range m.FSLayers {
		layers[i] = layer.BlobSum
	}
	return layers
}

func (i *dockerImage) getManifest() (manifest, error) {
	pr, err := ping(i.registry)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", pr.scheme+"://"+i.registry+"/v2/"+i.ref.RemoteName()+"/manifests/"+i.tag, nil)
	fmt.Println(req.URL.String())
	if err != nil {
		return nil, err
	}
	// TODO(runcom) set manifest version! schema1 for now - then schema2 etc etc and v1
	req.Header.Set("Docker-Distribution-API-Version", "registry/2.0")
	if pr.needsAuth() {
		req.SetBasicAuth(i.username, i.password)
		// support Docker bearer and abstract makeRequest
	}
	// insecure by default for now
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		// print body also
		return nil, fmt.Errorf("Invalid status code returned when fetching manifest %d", res.StatusCode)
	}
	manblob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	mschema1 := &manifestSchema1{}
	if err := json.Unmarshal(manblob, mschema1); err != nil {
		return nil, err
	}
	if err := fixManifestLayers(mschema1); err != nil {
		return nil, err
	}
	return mschema1, nil
}

// TODO(runcom): abstract makeRequest(req, headers)

func (i *dockerImage) GetLayers() ([]string, error) {
	m, err := i.getManifest()
	if err != nil {
		return nil, err
	}

	fmt.Println(m.GetLayers())

	return nil, nil
}

func parseDockerImage(img string) (Image, error) {
	ref, err := reference.ParseNamed(img)
	if err != nil {
		return nil, err
	}
	if reference.IsNameOnly(ref) {
		ref = reference.WithDefaultTag(ref)
	}
	var tag string
	switch x := ref.(type) {
	case reference.Canonical:
		tag = x.Digest().String()
	case reference.NamedTagged:
		tag = x.Tag()
	}
	var registry string
	hostname := ref.Hostname()
	if hostname == dockerHostname {
		registry = dockerRegistry
	} else {
		registry = hostname
	}
	username, password, err := getAuth(ref.Hostname())
	if err != nil {
		return nil, err
	}
	return &dockerImage{
		ref:      ref,
		tag:      tag,
		registry: registry,
		username: username,
		password: password,
	}, nil
}

func getAuth(hostname string) (string, string, error) {
	cfgFile, err := cliconfig.Load(cliconfig.ConfigDir())
	if err != nil {
		return "", "", err
	}
	if hostname == dockerHostname {
		hostname = dockerAuthRegistry
	}
	if auth, ok := cfgFile.AuthConfigs[hostname]; ok {
		return auth.Username, auth.Password, nil
	}
	return "", "", nil
}

type APIErr struct {
	Code    string
	Message string
	Detail  interface{}
}

type pingResponse struct {
	WWWAuthenticate string
	APIVersion      string
	scheme          string
	errors          []APIErr
}

func (pr *pingResponse) needsAuth() bool {
	return pr.WWWAuthenticate != ""
}

func ping(registry string) (*pingResponse, error) {
	// insecure by default for now
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	ping := func(scheme string) (*pingResponse, error) {
		resp, err := client.Get(scheme + "://" + registry + "/v2/")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
			return nil, fmt.Errorf("error pinging repository, response code %d", resp.StatusCode)
		}
		pr := &pingResponse{}
		pr.WWWAuthenticate = resp.Header.Get("WWW-Authenticate")
		pr.APIVersion = resp.Header.Get("Docker-Distribution-Api-Version")
		pr.scheme = scheme
		if resp.StatusCode == http.StatusUnauthorized {
			type APIErrors struct {
				Errors []APIErr
			}
			errs := &APIErrors{}
			if err := json.NewDecoder(resp.Body).Decode(errs); err != nil {
				return nil, err
			}
			pr.errors = errs.Errors
		}
		return pr, nil
	}
	scheme := "https"
	pr, err := ping(scheme)
	if err != nil {
		scheme = "http"
		pr, err = ping(scheme)
		if err == nil {
			return pr, nil
		}
	}
	return pr, err
}

func fixManifestLayers(manifest *manifestSchema1) error {
	type imageV1 struct {
		ID     string
		Parent string
	}
	imgs := make([]*imageV1, len(manifest.FSLayers))
	for i := range manifest.FSLayers {
		img := &imageV1{}

		if err := json.Unmarshal([]byte(manifest.History[i].V1Compatibility), img); err != nil {
			return err
		}

		imgs[i] = img
		if err := validateV1ID(img.ID); err != nil {
			return err
		}
	}
	if imgs[len(imgs)-1].Parent != "" {
		return errors.New("Invalid parent ID in the base layer of the image.")
	}
	// check general duplicates to error instead of a deadlock
	idmap := make(map[string]struct{})
	var lastID string
	for _, img := range imgs {
		// skip IDs that appear after each other, we handle those later
		if _, exists := idmap[img.ID]; img.ID != lastID && exists {
			return fmt.Errorf("ID %+v appears multiple times in manifest", img.ID)
		}
		lastID = img.ID
		idmap[lastID] = struct{}{}
	}
	// backwards loop so that we keep the remaining indexes after removing items
	for i := len(imgs) - 2; i >= 0; i-- {
		if imgs[i].ID == imgs[i+1].ID { // repeated ID. remove and continue
			manifest.FSLayers = append(manifest.FSLayers[:i], manifest.FSLayers[i+1:]...)
			manifest.History = append(manifest.History[:i], manifest.History[i+1:]...)
		} else if imgs[i].Parent != imgs[i+1].ID {
			return fmt.Errorf("Invalid parent ID. Expected %v, got %v.", imgs[i+1].ID, imgs[i].Parent)
		}
	}
	return nil
}

func validateV1ID(id string) error {
	if ok := validHex.MatchString(id); !ok {
		return fmt.Errorf("image ID %q is invalid", id)
	}
	return nil
}
