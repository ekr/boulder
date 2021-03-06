// Copyright 2015 ISRG.  All rights reserved
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// This mockgen call requires the import_prefix flag that is only present in the
// letsencrypt fork of mockgen.
// go:generate mockgen -source ./gsb.go -destination mock_gsb_test.go -package va -import_prefix github.com/letsencrypt/boulder/Godeps/_workspace/src

package va

import (
	safebrowsing "github.com/letsencrypt/boulder/Godeps/_workspace/src/github.com/letsencrypt/go-safe-browsing-api"
	"github.com/letsencrypt/boulder/core"
)

// SafeBrowsing is an interface for an third-party safe browing API client.
type SafeBrowsing interface {
	// IsListed returns a non-empty string if the domain was bad. Specifically,
	// that list is which Google Safe Browsing list the domain was found on.
	IsListed(url string) (list string, err error)
}

// IsSafeDomain returns true if the domain given is determined to be safe by an
// third-party safe browsing API. It's meant be called by the RA before pending
// authorization creation. If no third-party client was provided, it fails open
// and increments a Skips metric.
func (va *ValidationAuthorityImpl) IsSafeDomain(req *core.IsSafeDomainRequest) (*core.IsSafeDomainResponse, error) {
	va.stats.Inc("VA.IsSafeDomain.Requests", 1, 1.0)
	if va.SafeBrowsing == nil {
		va.stats.Inc("VA.IsSafeDomain.Skips", 1, 1.0)
		return &core.IsSafeDomainResponse{true}, nil
	}

	list, err := va.SafeBrowsing.IsListed(req.Domain)
	if err != nil {
		va.stats.Inc("VA.IsSafeDomain.Errors", 1, 1.0)
		if err == safebrowsing.ErrOutOfDateHashes {
			va.stats.Inc("VA.IsSafeDomain.OutOfDateHashErrors", 1, 1.0)
			return &core.IsSafeDomainResponse{true}, nil
		}
		return nil, err
	}
	va.stats.Inc("VA.IsSafeDomain.Successes", 1, 1.0)
	status := list == ""
	if status {
		va.stats.Inc("VA.IsSafeDomain.Status.Good", 1, 1.0)
	} else {
		va.stats.Inc("VA.IsSafeDomain.Status.Bad", 1, 1.0)
	}
	return &core.IsSafeDomainResponse{status}, nil
}
