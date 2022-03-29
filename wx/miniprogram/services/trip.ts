import { rental } from './proto-gen/rental/rental-pb'
import { Coolcar } from './request'

export namespace TripService {
  export async function create(
    req: rental.v1.ICreateTripRequest
  ): Promise<rental.v1.TripEntity> {
    return Coolcar.sendRequestWithAuthRetry({
      method: 'POST',
      path: '/v1/trips',
      data: rental.v1.CreateTripRequest.fromObject(req),
      respMarshaller: rental.v1.TripEntity.fromObject,
    })
  }

  export async function get(id: string): Promise<rental.v1.Trip> {
    return Coolcar.sendRequestWithAuthRetry({
      method: 'GET',
      path: `/v1/trips/${encodeURIComponent(id)}`,
      respMarshaller: rental.v1.Trip.fromObject,
    })
  }

  export async function list(
    status?: rental.v1.TripStatus
  ): Promise<rental.v1.GetTripsResponse> {
    let path = '/v1/trips'
    if (status) {
      path += `?status=${status}`
    }
    return Coolcar.sendRequestWithAuthRetry({
      method: 'GET',
      path,
      respMarshaller: rental.v1.GetTripsResponse.fromObject,
    })
  }
}
