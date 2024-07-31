package flags

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	ErrInvalidHost = errors.New("invalid host specified, must be either a HTTPS base URL (e.g., https://my.skuidsite.com) or a host (e.g., my.skuidsite.com)")
)

func ParseSince(val string) (string, error) {
	if val == "" {
		return val, nil
	}

	return util.GetTimestamp(val, time.Now(), true)
}

func ParseHost(rawUrl string) (string, error) {
	u, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		u, err = url.ParseRequestURI("https://" + rawUrl)
		if err != nil {
			return "", ErrInvalidHost
		}
	}

	if u.Scheme != "https" {
		return "", ErrInvalidHost
	}

	if u.Host == "" {
		return "", ErrInvalidHost
	} else {
		// TODO: Skuid Review Required - Not sure what the valid values here should be.  For example,
		// a custom domain, a multi-segment subdomain, port number, always *.skuidsite.com, etc.
		// This code should be modified (as needed) to ensure it validates expected values.
		// For now, validation is at lowest common viable option of domain.tld.
		if len(strings.Split(u.Hostname(), ".")) < 2 {
			return "", ErrInvalidHost
		}
	}

	// TODO: Skuid Review Required - Not sure what the valid values here should be.  For example,
	// are non-root paths supported in *.skuidsite.com?  in a custom domain? For now, only root
	// path is permitted
	if u.Path != "" && u.Path != "/" {
		return "", ErrInvalidHost
	}

	if u.RawQuery != "" {
		return "", ErrInvalidHost
	}

	// ParseRequestURI expects no fragment so if any exists, its treated as part of the path/query
	// therefore we would fail above in path or query check.  Fragment should always be "" at this point
	// given how ParseRequestURI behaves but just to be sure.
	if u.Fragment != "" {
		return "", ErrInvalidHost
	}

	// TODO: Skuid Review Required - Assuming user:pass never expected???
	if u.User != nil {
		return "", ErrInvalidHost
	}

	return fmt.Sprintf("%v://%v", u.Scheme, u.Host), nil
}
