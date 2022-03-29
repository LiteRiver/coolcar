import { rental } from './proto-gen/rental/rental-pb'
import { Coolcar } from './request'

export namespace TripService {
  export async function create(req: rental.v1.ICreateTripRequest) {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'POST',
      path: '/v1/trips',
      data: rental.v1.CreateTripRequest.fromObject(req),
      respMarshaller: rental.v1.TripEntity.fromObject,
    })
  }

  export async function get(id: string) {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'GET',
      path: `/v1/trips/${encodeURIComponent(id)}`,
      respMarshaller: rental.v1.Trip.fromObject,
    })
  }

  export async function list(status?: rental.v1.TripStatus) {
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

  export async function updateTripPos(id: string, loc?: rental.v1.ILocation) {
    return await update({
      id,
      current: loc,
    })
  }

  export async function endTrip(id: string) {
    return await update({ id, endTrip: true })
  }

  async function update(req: rental.v1.IUpdateTripRequest) {
    if (!req.id) {
      throw 'trip Id is required'
    }

    return await Coolcar.sendRequestWithAuthRetry({
      method: 'PUT',
      path: `/v1/trips/${encodeURIComponent(req.id)}`,
      data: rental.v1.UpdateTripRequest.fromObject(req),
      respMarshaller: rental.v1.Trip.fromObject,
    })
  }
}
