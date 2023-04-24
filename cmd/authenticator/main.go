package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
    "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"

    "github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func main() {
    log.Info(log.CAKC048, authenticator.FullVersionName)

    config, err := config.NewConfigFromEnv()
    if err != nil {
        printErrorAndExit(log.CAKC018)
    }

    authn, err := authenticator.NewAuthenticator(config)
    if err != nil {
        printErrorAndExit(log.CAKC019)
    }

    err = authn.AuthenticateWithContext(context.Background())
    if err != nil {
        printErrorAndExit(log.CAKC016)
    }

    fmt.Println("Authentication successful")

    os.Exit(0)
}

func printErrorAndExit(errorMessage string) {
    log.Error(errorMessage)
    os.Exit(1)
}

