package ai

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	coolenvpb "coolcar/shared/coolenv"
)

type Client struct {
	AIClient coolenvpb.AIServiceClient
}

func (c *Client) DistanceKm(ctx context.Context, from *rentalpb.Location, to *rentalpb.Location) (float64, error) {
	res, err := c.AIClient.MeasureDistance(ctx, &coolenvpb.MeasureDistanceRequest{
		From: &coolenvpb.Location{
			Latitude:  from.Latitude,
			Longitude: from.Longitude,
		},
		To: &coolenvpb.Location{
			Latitude:  to.Latitude,
			Longitude: to.Longitude,
		},
	})

	if err != nil {
		return 0, nil
	}

	return res.DistanceKm, nil
}
