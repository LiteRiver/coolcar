package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Manager struct {
	Provider
}

type Provider interface {
	GetProfile(context.Context, *rentalpb.GetProfileRequest) (*rentalpb.Profile, error)
}

func (p *Manager) Verify(ctx context.Context, accountId id.AccountId) (id.IdentityId, error) {
	nilIdentityId := id.IdentityId("")

	profile, err := p.Provider.GetProfile(ctx, &rentalpb.GetProfileRequest{})
	if err != nil {
		return nilIdentityId, fmt.Errorf("cannot get profile: %v", err)
	}

	if profile.IdentityStatus != rentalpb.IdentityStatus_VERIFIED {
		return nilIdentityId, fmt.Errorf("invalid identity status")
	}

	b, err := proto.Marshal(profile.Identity)
	if err != nil {
		return nilIdentityId, fmt.Errorf("cannot marshal identity: %v", err)
	}

	return id.IdentityId(base64.StdEncoding.EncodeToString(b)), nil
}
