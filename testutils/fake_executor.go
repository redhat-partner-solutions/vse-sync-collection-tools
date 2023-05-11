// SPDX-License-Identifier: GPL-2.0-or-later

package testutils

import (
	"context"
	"fmt"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Return a SPDYExectuor with stdout, stderr and an error embedded
func NewFakeNewSPDYExecutor(
	responder func(method string, url *url.URL) ([]byte, []byte, error),
	execCreationErr error,
) func(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	return func(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
		return &fakeExecutor{method: method, url: url, responder: responder}, execCreationErr
	}
}

type fakeExecutor struct {
	url       *url.URL
	responder func(method string, url *url.URL) ([]byte, []byte, error)
	method    string
}

func (f *fakeExecutor) Stream(options remotecommand.StreamOptions) error {
	stdout, stderr, responseErr := f.responder(f.method, f.url)
	_, err := options.Stdout.Write(stdout)
	if err != nil {
		return fmt.Errorf("failed to write stdout Error: %w", err)
	}
	_, err = options.Stderr.Write(stderr)
	if err != nil {
		return fmt.Errorf("failed to write stderr Error: %w", err)
	}
	return responseErr
}

func (f *fakeExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	stdout, stderr, reponseErr := f.responder(f.method, f.url)
	_, err := options.Stdout.Write(stdout)
	if err != nil {
		return fmt.Errorf("failed to write stdout Error: %w", err)
	}
	_, err = options.Stderr.Write(stderr)
	if err != nil {
		return fmt.Errorf("failed to write stderr Error: %w", err)
	}
	return reponseErr
}
