package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-opentelemetry-tracer/pkg/trace"
)

func main() {
	log.Info(log.CAKC048, authenticator.FullVersionName)

	var err error

	config, err := config.NewConfigFromEnv()
	if err != nil {
		printErrorAndExit(log.CAKC018)
	}

	tracer, _ := trace.NewTracerProvider(trace.NoopProviderType, false, trace.TracerProviderConfig{})
	defer tracer.Shutdown(context.Background())

	// Create new Authenticator
	authn, err := authenticator.NewAuthenticator(config)
	if err != nil {
		printErrorAndExit(log.CAKC019)
	}

	// Configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 1 * time.Second
	expBackoff.RandomizationFactor = 0
	expBackoff.Multiplier = 1
	expBackoff.MaxInterval = 2 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	err = backoff.Retry(func() error {
		for {
			err := authn.AuthenticateWithContext(context.Background())
			if err != nil {
				return log.RecordedError(log.CAKC016)
			}

			if config.GetContainerMode() == "init" {
				os.Exit(0)
			}

			log.Info(log.CAKC047, config.GetTokenTimeout())

			fmt.Println()
			time.Sleep(config.GetTokenTimeout())

			// Reset exponential backoff
			expBackoff.Reset()
		}
	}, expBackoff)

	if err != nil {
		printErrorAndExit(log.CAKC031)
	}
}

func printErrorAndExit(errorMessage string) {
	log.Error(errorMessage)
	os.Exit(0)
}
