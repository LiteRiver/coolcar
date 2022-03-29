export namespace routing {
  export interface DrivingOpts {
    tripId: string
  }

  export function driving(opts: DrivingOpts) {
    return `/pages/driving/driving?tripId=${opts.tripId}`
  }

  export interface UnlockOpts {
    carId: string
  }

  export function unlock(opts: UnlockOpts) {
    return `/pages/unlock/unlock?carId=${opts.carId}`
  }

  export interface RegistrationOpts {
    redirect?: string
  }

  export interface RegistartionParams {
    redirectURL: string
  }

  export function registration(params?: RegistartionParams) {
    const page = `/pages/registration/registration`
    if (!params) {
      return page
    }

    return `${page}?redirect=${encodeURIComponent(params.redirectURL)}`
  }

  export function mytrips() {
    return '/pages/mytrips/mytrips'
  }
}
