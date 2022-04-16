import camelcaseKeys from 'camelcase-keys'
import { car } from './proto-gen/car/car-pb'
import { Coolcar } from './request'

export namespace CarService {
  export function subscribe(onMsg: (c: car.v1.ICarEntity) => void) {
    const socket = wx.connectSocket({
      url: `${Coolcar.WS_ADDR}/ws`,
    })
    socket.onMessage((msg) => {
      const obj = JSON.parse(msg.data as string)
      onMsg(car.v1.CarEntity.fromObject(camelcaseKeys(obj, { deep: true })))
    })
    return socket
  }

  export function get(id: string) {
    return Coolcar.sendRequestWithAuthRetry({
      method: 'GET',
      path: `/v1/cars/${encodeURIComponent(id)}`,
      respMarshaller: car.v1.Car.fromObject,
    })
  }
}
