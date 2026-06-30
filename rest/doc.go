// Package rest provides an HTTP server and client for RESTful services.
//
// # Relationship to turbo
//
// rest is the service-level layer; turbo is the bare router. rest/Server
// wraps a turbo.Router and adds:
//
//   - lifecycle integration (registers as a lifecycle.Component, so
//     startup/shutdown is coordinated with the rest of the application)
//   - TLS configuration (cert/key paths, min version, optional custom
//     tls.Config)
//   - content-negotiated bind / respond via golly's codec package
//     (JSON / XML / YAML, picked by Content-Type and Accept)
//   - PathPrefix option (prefixes every registered route)
//   - opinionated default options (DefaultSrvOptions / Default)
//   - optional access-log + panic-recovery middleware
//
// Choose rest when you want a service framework — lifecycle, TLS,
// codec wiring, observability hooks. Choose turbo directly when you
// want to embed a router in something that already owns the
// http.Server (a binary that serves multiple protocols, a test
// harness, etc.). Anything you register on rest is registered on
// turbo underneath, so the two are not divergent — they layer.
//
// # Middleware
//
// Middleware is turbo.FilterFunc — a func(http.Handler) http.Handler.
// Three places to attach one:
//
//   - Server.AddGlobalFilter(f) — runs for every request on the server
//   - the returned *turbo.Route from Get/Post/... — route.AddFilter(f)
//     runs only for that route
//   - turbo.Group (via Server.Router().Group("/prefix", f...)) — applies
//     to all routes registered on the group
//
// Within one route the filter order is: authenticator (if any) → global
// filters in registration order → route filters in registration order
// → handler.
//
// Ready-to-use middleware shipped with the package:
//
//   - rest.AccessLog() — logs method, path, status, latency, bytes
//     via golly/l3
//   - rest.Recover() — recovers panics, writes 500, logs with stack
//
// # Quick start
//
//	srv, _ := rest.DefaultServer()
//	srv.Opts().PathPrefix = "/api/v1"
//	srv.AddGlobalFilter(rest.Recover())
//	srv.AddGlobalFilter(rest.AccessLog())
//	srv.Get("/users/:id", func(ctx rest.ServerContext) {
//	    var u User
//	    if err := ctx.Bind(&u); err != nil { ctx.SetStatusCode(400); return }
//	    _ = ctx.Respond(200, u) // content-negotiated by Accept
//	})
//	manager := lifecycle.NewSimpleComponentManager()
//	manager.Register(srv); manager.StartAndWait()
//
// # Client
//
// rest also exposes a Client. Authentication is delegated to the
// golly/clients package (Basic / Bearer / API key); retry and circuit-
// breaker integration is in progress — see the package source for
// current status.
package rest
