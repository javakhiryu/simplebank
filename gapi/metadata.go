package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	grpcGatewayIpAddressHeader = "x-forwarded-for"
	grpcUserAgentHeader        = "grpc-client"
)

type Metadata struct {
	UserAgent string
	IpAddress string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if userAgents := md.Get(grpcUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if IpAddresses := md.Get(grpcGatewayIpAddressHeader); len(IpAddresses) > 0 {
			mtdt.IpAddress = IpAddresses[0]
		}
		if p, ok := peer.FromContext(ctx); ok {
			mtdt.IpAddress = p.Addr.String()
		}

	}
	return mtdt
}
