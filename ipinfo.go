package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bouk/httprouter"
	"github.com/cloudflare/golibs/lrucache"
	"github.com/phuslu/glog"
	"golang.org/x/sync/singleflight"
)

type IpinfoHandler struct {
	URL          string
	Regex        *regexp.Regexp
	Cache        lrucache.Cache
	CacheTTL     time.Duration
	Singleflight *singleflight.Group
	Transport    *http.Transport
}

type IpinfoResponse struct {
	Error    string `json:"error,omitempty""`
	Location string `json:"location,omitempty"`
	ISP      string `json:"isp,omitempty"`
}

func (h *IpinfoHandler) Error(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(IpinfoResponse{
		Error: err.Error(),
	})
}

func (h *IpinfoHandler) Ipinfo(w http.ResponseWriter, r *http.Request) {
	glog.Infof("%s \"%s %s\" \"%s\"", r.RemoteAddr, r.Method, r.URL, r.Header.Get("User-Agent"))

	var err error
	var item *IpinfoItem

	ipStr := httprouter.GetParam(r, "ip")
	if ipStr == "" {
		ipStr, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	key := "ipinfo:" + ipStr
	if v, ok := h.Cache.GetNotStale(key); ok {
		item = v.(*IpinfoItem)
	} else {
		item, err = h.ipinfoSearch(ipStr)
		if err != nil {
			h.Error(w, err)
			return
		}

		h.Cache.Set(key, item, time.Now().Add(h.CacheTTL))
	}

	json.NewEncoder(w).Encode(IpinfoResponse{
		Error:    "",
		Location: item.Location,
		ISP:      item.ISP,
	})
}

type IpinfoItem struct {
	Location string
	ISP      string
}

func (h *IpinfoHandler) ipinfoSearch(ipStr string) (*IpinfoItem, error) {
	url := strings.Replace(h.URL, "%s", ipStr, 1)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "curl/7.56.0")

	v, err, _ := h.Singleflight.Do(url, func() (interface{}, error) {
		return h.Transport.RoundTrip(req)
	})
	if err != nil {
		return nil, err
	}

	resp := v.(*http.Response)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	match := h.Regex.FindStringSubmatch(string(data))
	if match == nil {
		return nil, fmt.Errorf("empty")
	}

	item := &IpinfoItem{
		Location: match[1],
		ISP:      match[2],
	}

	glog.Infof("ipinfoSearch(%#v) return %+v", ipStr, item)

	h.Singleflight.Forget(url)

	return item, nil
}
