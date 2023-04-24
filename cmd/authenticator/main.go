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
    expBackoff.MaxInterval = 1 * time.Second
    expBackoff.MaxElapsedTime = 5 * time.Second

    retryCount := 0

    _ = backoff.Retry(func() error {
        retryCount++
        err := authn.AuthenticateWithContext(context.Background())
        if err != nil {
            if retryCount >= 2 {
                return nil
            }
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

        return nil
    }, expBackoff)

    os.Exit(0)
}

func printErrorAndExit(errorMessage string) {
    log.Error(errorMessage)
    os.Exit(1)
}