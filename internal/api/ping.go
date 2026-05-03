package api

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	adminv1 "website-admin/gen/admin/v1"
	"website-admin/gen/admin/v1/adminv1connect"
	"website-admin/internal/jot"
)

type PingServer struct{}

func (s *PingServer) AdminPing(
	ctx context.Context,
	req *connect.Request[adminv1.AdminPingRequest],
) (*connect.Response[adminv1.AdminPingResponse], error) {
	jot.Info("Deep Admin Server received: " + req.Msg.TestValue)

	return connect.NewResponse(&adminv1.AdminPingResponse{
		TestValue: req.Msg.TestValue,
		Message:   "Hello from the secure website-admin backend!",
	}), nil
}

func Register(mux *http.ServeMux) {
	path, handler := adminv1connect.NewAdminServiceHandler(&PingServer{})

	mux.Handle(path, handler)

	jot.Info("Registered Connect-RPC API at path: " + path)
}
