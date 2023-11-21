/*
 * Created by lintao on 2023/7/26 下午2:22
 * Copyright © 2020-2023 LINTAO. All rights reserved.
 *
 */

package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/NSObjects/echo-admin/internal/api/data"
	"github.com/NSObjects/echo-admin/internal/api/service"
	"github.com/NSObjects/echo-admin/internal/code"
	"github.com/NSObjects/echo-admin/internal/configs"
	"github.com/NSObjects/echo-admin/internal/log"
	"github.com/NSObjects/echo-admin/internal/resp"
	"github.com/NSObjects/echo-admin/internal/server/middlewares"
	"github.com/casbin/casbin/v2"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	casbin_mw "github.com/labstack/echo-contrib/casbin"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marmotedu/errors"
)

type EchoServer struct {
	server  *echo.Echo
	Routers []service.RegisterRouter `group:"routes"`
	cfg     configs.Config
}

func (s *EchoServer) Server() *echo.Echo {
	return s.server
}

func NewEchoServer(routes []service.RegisterRouter, c *casbin.Enforcer, cfg configs.Config) *EchoServer {
	s := &EchoServer{
		server:  echo.New(),
		Routers: routes,
		cfg:     cfg,
	}
	s.loadMiddleware(c)
	s.registerRouter()
	return s
}

func errorHandler(err error, c echo.Context) {
	er := resp.APIError(err, c)
	if er != nil {
		log.Error(er)
	}
}

func (s *EchoServer) loadMiddleware(c *casbin.Enforcer) {
	s.server.Validator = &middlewares.Validator{Validator: validator.New()}
	s.server.Use(middleware.Gzip())
	s.server.HTTPErrorHandler = errorHandler

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(data.JwtCustomClaims)
		},
		SigningKey: []byte(s.cfg.JWT.Secret),
		Skipper: func(c echo.Context) bool {
			return c.Request().RequestURI == "/api/login"
		},
	}

	s.server.Use(echojwt.WithConfig(config))
	//s.server.Use(middleware.Recover())
	s.server.Use(casbin_mw.MiddlewareWithConfig(casbin_mw.Config{
		Enforcer: c,
		Skipper: func(c echo.Context) bool {
			return c.Request().RequestURI == "/api/login"
		},
		ErrorHandler: func(c echo.Context, internal error, proposedStatus int) error {
			return errors.WrapC(internal, code.ErrPermissionDenied, "权限不足")
		},
		UserGetter: func(c echo.Context) (string, error) {
			token := c.Get("user").(*jwt.Token)
			if token == nil {
				return "", nil
			}

			user := token.Claims.(*data.JwtCustomClaims)
			if user == nil {
				return "", nil
			}
			if user.Admin {
				return "root", nil
			}

			return user.Name, nil
		},
	}))

	s.server.Use(middleware.Logger())
	s.server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		//todo 域名设置
		//AllowOrigins:     []string{"http://xxx:8080","https://xxxx:8080"},
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowCredentials: true,
	}))
}

func (s *EchoServer) registerRouter() {
	g := s.server.Group("api")
	for _, v := range s.Routers {
		v.RegisterRouter(g)
	}
}

func (s *EchoServer) Run(port string) {
	go func() {
		if err := s.server.Start(port); err != nil && err != http.ErrServerClosed {
			s.server.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:gomnd
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.server.Logger.Fatal(err)
	}
}
