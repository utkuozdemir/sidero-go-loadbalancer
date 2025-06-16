// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package loadbalancer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/siderolabs/go-loadbalancer/upstream"
)

type Node struct {
	logger  *zap.Logger
	Address string // host:port
}

func (upstream Node) HealthCheck(ctx context.Context) (upstream.Tier, error) {
	start := time.Now()
	err := upstream.healthCheck(ctx)
	elapsed := time.Since(start)

	return calcTier(err, elapsed)
}

func (upstream Node) healthCheck(ctx context.Context) error {
	d := probeDialer()

	c, err := d.DialContext(ctx, "tcp", upstream.Address)
	if err != nil {
		upstream.logger.Warn("healthcheck failed", zap.String("address", upstream.Address), zap.Error(err))

		return err
	}

	return c.Close()
}

var mins = []time.Duration{
	0,
	time.Millisecond / 10,
	time.Millisecond,
	10 * time.Millisecond,
	100 * time.Millisecond,
	1 * time.Second,
}

func calcTier(err error, elapsed time.Duration) (upstream.Tier, error) {
	if err != nil {
		// preserve old tier
		return -1, err
	}

	for i := len(mins) - 1; i >= 0; i-- {
		if elapsed >= mins[i] {
			return upstream.Tier(i), nil
		}
	}

	// We should never get here, but there is no way to tell this to Go compiler.
	return upstream.Tier(len(mins)), err
}
