import { rental } from './proto-gen/rental/rental-pb'
import { Coolcar } from './request'

export namespace TripService {
  export async function create(
    req: rental.v1.CreateTripRequest
  ): Promise<rental.v1.CreateTripResponse> {
    return Coolcar.sendRequestWithAuthRetry({
      method: 'POST',
      path: '/v1/rental/trip',
      data: req,
      respMarshaller: rental.v1.CreateTripResponse.fromObject,
    })
  }
}
