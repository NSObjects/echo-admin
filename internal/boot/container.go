package boot

import (
	"context"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"

	"github.com/NSObjects/echo-admin/internal/platform/configs"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/logging"
	"github.com/NSObjects/echo-admin/internal/platform/infrastructure/resources"
	"github.com/NSObjects/echo-admin/internal/platform/server"
)

type loggingReady struct{}

func newInjector(ctx context.Context, cfg configs.Config) do.Injector {
	if ctx == nil {
		ctx = context.Background()
	}
	return do.New(
		provideStartupContext(ctx),
		provideConfig(cfg),
		provideLogging,
		provideResources(ctx),
		provideResourceClients,
		provideServer,
	)
}

func provideStartupContext(ctx context.Context) func(do.Injector) {
	return func(i do.Injector) {
		do.Provide(i, func(do.Injector) (context.Context, error) {
			return ctx, nil
		})
	}
}

func provideConfig(cfg configs.Config) func(do.Injector) {
	return func(i do.Injector) {
		do.Provide(i, func(do.Injector) (configs.Config, error) {
			normalized := configs.Normalize(cfg)
			if err := configs.Validate(normalized); err != nil {
				return configs.Config{}, fmt.Errorf("validate config: %w", err)
			}
			return normalized, nil
		})
	}
}

func provideLogging(i do.Injector) {
	do.Provide(i, func(i do.Injector) (loggingReady, error) {
		cfg, err := do.Invoke[configs.Config](i)
		if err != nil {
			return loggingReady{}, err
		}
		if err := logging.Install(logging.FromAppConfig(cfg)); err != nil {
			return loggingReady{}, fmt.Errorf("configure logging: %w", err)
		}
		return loggingReady{}, nil
	})
}

func provideResources(ctx context.Context) func(do.Injector) {
	return func(i do.Injector) {
		do.Provide(i, func(i do.Injector) (*Resources, error) {
			if _, err := do.Invoke[loggingReady](i); err != nil {
				return nil, err
			}
			cfg, err := do.Invoke[configs.Config](i)
			if err != nil {
				return nil, err
			}
			return openResources(ctx, cfg)
		})
	}
}

func provideResourceClients(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*gorm.DB, error) {
		bundle, err := do.Invoke[*Resources](i)
		if err != nil {
			return nil, err
		}
		db, ok := bundle.MySQL()
		if !ok {
			return nil, disabledResourceError(resources.CapabilityMySQL)
		}
		return db, nil
	})

	do.Provide(i, func(i do.Injector) (*goredis.Client, error) {
		bundle, err := do.Invoke[*Resources](i)
		if err != nil {
			return nil, err
		}
		client, ok := bundle.Redis()
		if !ok {
			return nil, disabledResourceError(resources.CapabilityRedis)
		}
		return client, nil
	})

	do.Provide(i, func(i do.Injector) (*mongo.Client, error) {
		bundle, err := do.Invoke[*Resources](i)
		if err != nil {
			return nil, err
		}
		client, ok := bundle.MongoDB()
		if !ok {
			return nil, disabledResourceError(resources.CapabilityMongoDB)
		}
		return client, nil
	})
}

func provideServer(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*server.Server, error) {
		cfg, err := do.Invoke[configs.Config](i)
		if err != nil {
			return nil, err
		}
		resourceBundle, err := do.Invoke[*Resources](i)
		if err != nil {
			return nil, err
		}
		options := []server.Option{server.WithStatusReporter(resourceBundle)}
		options, err = appendOptionalAPIKeyVerifier(i, options)
		if err != nil {
			return nil, err
		}
		options, err = appendOptionalSystemErrorRecorder(i, options)
		if err != nil {
			return nil, err
		}
		options, err = appendOptionalLoginSessionAuthenticator(i, options)
		if err != nil {
			return nil, err
		}
		srv, err := server.New(cfg, options...)
		if err != nil {
			return nil, fmt.Errorf("create server: %w", err)
		}
		return srv, nil
	})
}

func appendOptionalAPIKeyVerifier(i do.Injector, options []server.Option) ([]server.Option, error) {
	verifier, err := do.InvokeAs[server.APIKeyVerifier](i)
	if err == nil {
		return append(options, server.WithAPIKeyVerifier(verifier)), nil
	}
	if optionalServiceMissing(err) {
		return options, nil
	}
	return nil, err
}

func appendOptionalSystemErrorRecorder(i do.Injector, options []server.Option) ([]server.Option, error) {
	recorder, err := do.InvokeAs[server.SystemErrorRecorder](i)
	if err == nil {
		return append(options, server.WithSystemErrorRecorder(recorder)), nil
	}
	if optionalServiceMissing(err) {
		return options, nil
	}
	return nil, err
}

func appendOptionalLoginSessionAuthenticator(i do.Injector, options []server.Option) ([]server.Option, error) {
	authenticator, err := do.InvokeAs[server.LoginSessionAuthenticator](i)
	if err == nil {
		return append(options, server.WithLoginSessionAuthenticator(authenticator)), nil
	}
	if optionalServiceMissing(err) {
		return options, nil
	}
	return nil, err
}

func optionalServiceMissing(err error) bool {
	return errors.Is(err, do.ErrServiceNotFound) || errors.Is(err, do.ErrServiceNotMatch)
}

func disabledResourceError(capability string) error {
	return resources.NewCapabilityError(capability, "inject", errors.New("resource is disabled"))
}
