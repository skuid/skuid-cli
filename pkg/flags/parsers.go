package flags

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	ErrInvalidHost = errors.New("invalid host specified, must be either a HTTPS base URL (e.g., https://my.skuidsite.com) or a host (e.g., my.skuidsite.com)")
)

type ParseStringOptions struct {
	AllowBlank bool
}

func ParseSince(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}

	t, err := util.GetTime(val, time.Now(), true)
	if err != nil {
		return nil, err
	}
	return &t, err
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

func ParseString(options ParseStringOptions) func(val string) (string, error) {
	return func(val string) (string, error) {
		if !options.AllowBlank && strings.TrimSpace(val) == "" {
			return "", fmt.Errorf("must contain a non-blank value")
		}
		return val, nil
	}
}

func ParseLogLevel(val string) (logrus.Level, error) {
	// simple wrapper to customize error message, otherwise it will contain the word logrus
	if l, err := logrus.ParseLevel(strings.TrimSpace(val)); err != nil {
		return l, fmt.Errorf("not a valid log level: %v", logging.QuoteText(val))
	} else {
		return l, nil
	}
}

func ParseMetadataEntity(val string) (metadata.MetadataEntity, error) {
	if e, err := metadata.NewMetadataEntity(strings.TrimSpace(val)); err != nil {
		return metadata.MetadataEntity{}, err
	} else {
		return *e, nil
	}
}

func ParseDirectory(val string) (string, error) {
	if targetDirectory, err := filepath.Abs(val); err != nil {
		return "", fmt.Errorf("unable to convert %v to absolute path: %w", logging.QuoteText(val), err)
	} else {
		return targetDirectory, nil
	}
}
