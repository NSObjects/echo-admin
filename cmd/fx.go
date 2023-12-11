/*
 * Created by lintao on 2023/7/27 上午10:04
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package cmd

import (
	"context"
	"github.com/NSObjects/echo-admin/internal/api/biz"
	"github.com/NSObjects/echo-admin/internal/api/data"
	"github.com/NSObjects/echo-admin/internal/api/service"
	"github.com/NSObjects/echo-admin/internal/log"
	"github.com/marmotedu/errors"
	"go.uber.org/fx"
	"log/slog"

	"github.com/NSObjects/echo-admin/internal/configs"
	"github.com/NSObjects/echo-admin/internal/server"
)

func Run(cfg string) {
	fx.New(
		fx.Provide(func() configs.Config {
			config := configs.NewCfg(cfg)
			//log.Init(config)
			log.Error(errors.New("test"), slog.String("xxx", "xxx"))
			return config
		}),
		data.Model,
		biz.Model,
		service.Model,
		fx.Provide(
			fx.Annotate(
				server.NewEchoServer,
				fx.ParamTags(`group:"routes"`)),
		),
		fx.Invoke(func(lifecycle fx.Lifecycle, s *server.EchoServer, cfg configs.Config) {
			lifecycle.Append(
				fx.Hook{
					OnStart: func(context.Context) error {
						go s.Run(cfg.System.Port)
						return nil
					},
				})
		})).Run()
}
