import { rental } from './proto-gen/rental/rental-pb'
import { Coolcar } from './request'

export namespace ProfileService {
  export async function get() {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'GET',
      path: '/v1/profile',
      respMarshaller: rental.v1.Profile.fromObject,
    })
  }

  export async function submit(req: rental.v1.IIdentity) {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'POST',
      path: '/v1/profile',
      data: rental.v1.Identity.fromObject(req),
      respMarshaller: rental.v1.Profile.fromObject,
    })
  }

  export async function clear() {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'DELETE',
      path: '/v1/profile',
      respMarshaller: rental.v1.Profile.fromObject,
    })
  }

  export async function getPhoto() {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'GET',
      path: '/v1/profile/photo',
      respMarshaller: rental.v1.GetProfilePhotoResponse.fromObject,
    })
  }

  export async function createPhoto() {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'POST',
      path: '/v1/profile/photo',
      respMarshaller: rental.v1.CreateProfilePhotoResponse.fromObject,
    })
  }

  export async function completePhoto() {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'POST',
      path: '/v1/profile/photo/complete',
      respMarshaller: rental.v1.Identity.fromObject,
    })
  }

  export async function clearPhoto() {
    return await Coolcar.sendRequestWithAuthRetry({
      method: 'DELETE',
      path: '/v1/profile/photo',
      respMarshaller: rental.v1.ClearProfilePhotoResponse.fromObject,
    })
  }
}
